package collector

import (
	"context"
	"sort"

	"github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	"github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/giantswarm/etcd-backup-operator/v5/service/controller/key"
)

const (
	labelTenantClusterId = "tenant_cluster_id"
	labelETCDVersion     = "etcd_version"

	backupStateCompleted = "Completed"
	backupStateSkipped   = "Skipped"
)

var (
	namespace = "etcd_backup"
	labels    = []string{labelTenantClusterId, labelETCDVersion}

	creationTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "creation_time_ms"),
		"Gauge about the time in ms spent by the ETCD backup creation process.",
		labels,
		nil,
	)

	encryptionTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "encryption_time_ms"),
		"Gauge about the time in ms spent by the ETCD backup encryption process.",
		labels,
		nil,
	)

	uploadTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "upload_time_ms"),
		"Gauge about the time in ms spent by the ETCD backup upload process.",
		labels,
		nil,
	)

	backupSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "size_bytes"),
		"Gauge about the size of the backup file, as seen by S3.",
		labels,
		nil,
	)

	latestAttemptTimestampDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "latest_attempt"),
		"Timestamp of the latest backup attempt",
		labels,
		nil,
	)

	latestSuccessTimestampDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "latest_success"),
		"Timestamp of the latest backup succeeded",
		labels,
		nil,
	)
)

type ETCDBackupConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type ETCDBackup struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewETCDBackup(config ETCDBackupConfig) (*ETCDBackup, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	d := &ETCDBackup{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return d, nil
}

func (d *ETCDBackup) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get a list of all ETCDBackup objects.
	backupListResult := v1alpha1.ETCDBackupList{}
	err := d.k8sClient.CtrlClient().List(ctx, &backupListResult)
	if err != nil {
		return microerror.Mask(err)
	}

	// Sort backups by Status.FinishedTimestamp.
	backups := backupListResult.Items
	sort.Slice(backups, func(i, j int) bool {
		t1 := backups[i].Status.FinishedTimestamp.Time
		t2 := backups[j].Status.FinishedTimestamp.Time
		return t1.Before(t2)
	})

	// Get a list of current tenant clusters.
	tenantClusterIds, err := d.getTenantClusterIDs(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantClusterIds = append(tenantClusterIds, key.ManagementCluster)

	// Iterate over all ETCDBackup objects and select the most recent backup from each cluster.
	latestV3SuccessMetrics := map[string]v1alpha1.ETCDInstanceBackupStatus{}
	latestV3AttemptMetrics := map[string]v1alpha1.ETCDInstanceBackupStatus{}

	for _, backup := range backups {
		for _, instanceStatus := range backup.Status.Instances {
			// The cluster this instance status is referring to does not exist anymore.
			// We simply ignore this metric otherwise we get paged for deleted clusters.
			if !inSlice(instanceStatus.Name, tenantClusterIds) {
				continue
			}

			if instanceStatus.V3 != nil && instanceStatus.V3.Status == backupStateCompleted {
				latestV3SuccessMetrics[instanceStatus.Name] = *instanceStatus.V3
			}

			if instanceStatus.V3 != nil && instanceStatus.V3.Status != backupStateSkipped {
				latestV3AttemptMetrics[instanceStatus.Name] = *instanceStatus.V3
			}
		}
	}

	sendAttemptMetricsForVersion := func(tenantClusterID string, status v1alpha1.ETCDInstanceBackupStatus, version string) {
		ch <- prometheus.MustNewConstMetric(
			latestAttemptTimestampDesc,
			prometheus.GaugeValue,
			float64(status.FinishedTimestamp.Unix()),
			tenantClusterID,
			version,
		)
	}

	sendSuccessMetricsForVersion := func(tenantClusterID string, status v1alpha1.ETCDInstanceBackupStatus, version string) {
		if status.Status == backupStateCompleted {
			ch <- prometheus.MustNewConstMetric(
				creationTimeDesc,
				prometheus.GaugeValue,
				float64(status.CreationTime),
				tenantClusterID,
				version,
			)

			ch <- prometheus.MustNewConstMetric(
				encryptionTimeDesc,
				prometheus.GaugeValue,
				float64(status.EncryptionTime),
				tenantClusterID,
				version,
			)

			ch <- prometheus.MustNewConstMetric(
				uploadTimeDesc,
				prometheus.GaugeValue,
				float64(status.UploadTime),
				tenantClusterID,
				version,
			)

			ch <- prometheus.MustNewConstMetric(
				backupSizeDesc,
				prometheus.GaugeValue,
				float64(status.BackupFileSize),
				tenantClusterID,
				version,
			)

			ch <- prometheus.MustNewConstMetric(
				latestSuccessTimestampDesc,
				prometheus.GaugeValue,
				float64(status.FinishedTimestamp.Unix()),
				tenantClusterID,
				version,
			)
		}
	}

	for clusterName, status := range latestV3SuccessMetrics {
		sendSuccessMetricsForVersion(clusterName, status, "V3")
	}

	for clusterName, status := range latestV3AttemptMetrics {
		sendAttemptMetricsForVersion(clusterName, status, "V3")
	}

	return nil
}

func (d *ETCDBackup) Describe(ch chan<- *prometheus.Desc) error {
	ch <- creationTimeDesc
	ch <- encryptionTimeDesc
	ch <- uploadTimeDesc
	ch <- backupSizeDesc
	ch <- latestAttemptTimestampDesc
	ch <- latestSuccessTimestampDesc
	return nil
}

func (d *ETCDBackup) getTenantClusterIDs(ctx context.Context) ([]string, error) {
	crdClient := d.k8sClient.CtrlClient()
	var ret []string

	// AWS
	{
		crdList := providerv1alpha1.AWSConfigList{}
		err := crdClient.List(ctx, &crdList)
		if err == nil {
			for _, awsConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if awsConfig.DeletionTimestamp.IsZero() {
					ret = append(ret, awsConfig.Name)
				}
			}
		}
	}

	// AWS cluster API
	{
		crdList := v1alpha3.AWSClusterList{}
		err := crdClient.List(ctx, &crdList)
		if err == nil {
			for _, awsClusterObj := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if awsClusterObj.DeletionTimestamp.IsZero() {
					ret = append(ret, awsClusterObj.Name)
				}
			}
		}
	}

	// Azure
	{
		crdList := providerv1alpha1.AzureConfigList{}
		err := crdClient.List(ctx, &crdList)
		if err == nil {
			for _, azureConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if azureConfig.DeletionTimestamp.IsZero() {
					ret = append(ret, azureConfig.Name)
				}
			}
		}
	}

	// KVM
	{
		crdList := providerv1alpha1.KVMConfigList{}
		err := crdClient.List(ctx, &crdList)
		if err == nil {
			for _, kvmConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if kvmConfig.DeletionTimestamp.IsZero() {
					ret = append(ret, kvmConfig.Name)
				}
			}
		}
	}

	// CAPI
	{
		crdList := capi.ClusterList{}
		err := crdClient.List(ctx, &crdList)
		if err == nil {
			for _, cluster := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if cluster.DeletionTimestamp == nil {
					ret = append(ret, cluster.Name)
				}
			}
		}
	}

	return ret, nil
}

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

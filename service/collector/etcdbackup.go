package collector

import (
	"sort"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
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
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type ETCDBackup struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func NewETCDBackup(config ETCDBackupConfig) (*ETCDBackup, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	d := &ETCDBackup{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return d, nil
}

func (d *ETCDBackup) Collect(ch chan<- prometheus.Metric) error {
	// Get a list of all ETCDBackup objects.
	backupListResult, err := d.g8sClient.BackupV1alpha1().ETCDBackups().List(v1.ListOptions{})
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
	tenantClusterIds, err := d.getTenantClusterIDs()
	if err != nil {
		return microerror.Mask(err)
	}

	tenantClusterIds = append(tenantClusterIds, key.ControlPlane)

	// Iterate over all ETCDBackup objects and select the most recent backup from each cluster.
	latestV2SuccessMetrics := map[string]v1alpha1.ETCDInstanceBackupStatus{}
	latestV3SuccessMetrics := map[string]v1alpha1.ETCDInstanceBackupStatus{}
	latestV2AttemptMetrics := map[string]v1alpha1.ETCDInstanceBackupStatus{}
	latestV3AttemptMetrics := map[string]v1alpha1.ETCDInstanceBackupStatus{}

	for _, backup := range backups {
		for _, instanceStatus := range backup.Status.Instances {
			// The cluster this instance status is referring to does not exist anymore.
			// We simply ignore this metric otherwise we get paged for deleted clusters.
			if !inSlice(instanceStatus.Name, tenantClusterIds) {
				continue
			}

			if instanceStatus.V2.Status == backupStateCompleted {
				latestV2SuccessMetrics[instanceStatus.Name] = instanceStatus.V2
			}
			if instanceStatus.V3.Status == backupStateCompleted {
				latestV3SuccessMetrics[instanceStatus.Name] = instanceStatus.V3
			}

			if instanceStatus.V2.Status != backupStateSkipped {
				latestV2AttemptMetrics[instanceStatus.Name] = instanceStatus.V2
			}
			if instanceStatus.V3.Status != backupStateSkipped {
				latestV3AttemptMetrics[instanceStatus.Name] = instanceStatus.V3
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

	for clusterName, status := range latestV2SuccessMetrics {
		sendSuccessMetricsForVersion(clusterName, status, "V2")
	}

	for clusterName, status := range latestV3SuccessMetrics {
		sendSuccessMetricsForVersion(clusterName, status, "V3")
	}

	for clusterName, status := range latestV2AttemptMetrics {
		sendAttemptMetricsForVersion(clusterName, status, "V2")
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

func (d *ETCDBackup) getTenantClusterIDs() ([]string, error) {
	crdClient := d.g8sClient
	var ret []string

	// AWS
	{
		crdList, err := crdClient.ProviderV1alpha1().AWSConfigs(v1.NamespaceAll).List(v1.ListOptions{})
		if err == nil {
			for _, awsConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if awsConfig.DeletionTimestamp == nil {
					ret = append(ret, awsConfig.Name)
				}
			}
		}
	}

	// AWS cluster API
	{
		crdList, err := crdClient.InfrastructureV1alpha2().AWSClusters(v1.NamespaceAll).List(v1.ListOptions{})
		if err == nil {
			for _, awsClusterObj := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if awsClusterObj.DeletionTimestamp == nil {
					ret = append(ret, awsClusterObj.Name)
				}
			}
		}
	}

	// Azure
	{
		crdList, err := crdClient.ProviderV1alpha1().AzureConfigs(v1.NamespaceAll).List(v1.ListOptions{})
		if err == nil {
			for _, azureConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if azureConfig.DeletionTimestamp == nil {
					ret = append(ret, azureConfig.Name)
				}
			}
		}
	}

	// KVM
	{
		crdList, err := crdClient.ProviderV1alpha1().KVMConfigs(v1.NamespaceAll).List(v1.ListOptions{})
		if err == nil {
			for _, kvmConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if kvmConfig.DeletionTimestamp == nil {
					ret = append(ret, kvmConfig.Name)
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

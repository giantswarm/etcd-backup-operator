package collector

import (
	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	labelTenantClusterId = "tenant_cluster_id"
	labelETCDVersion     = "etcd_version"

	backupStateCompleted = "Completed"
	backupStateFailed    = "Failed"
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

	attemptsCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "attempts_count"),
		"Count of attempted backups",
		labels,
		nil,
	)

	successCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "success_count"),
		"Count of successful backups",
		labels,
		nil,
	)

	failureCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "failure_count"),
		"Count of failed backups",
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

	// EnvironmentName is the name of the Azure environment used to compute the
	// azure.Environment type. See also
	// https://godoc.org/github.com/Azure/go-autorest/autorest/azure#Environment.
	EnvironmentName string
}

type ETCDBackup struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	environmentName string
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

	if config.EnvironmentName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EnvironmentName must not be empty", config)
	}

	d := &ETCDBackup{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		environmentName: config.EnvironmentName,
	}

	return d, nil
}

func (d *ETCDBackup) Collect(ch chan<- prometheus.Metric) error {
	// Get a list of all ETCDBackup objects.
	backups, err := d.g8sClient.BackupV1alpha1().ETCDBackups().List(v1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	var newest *v1alpha1.ETCDBackup

	// Find the most recent ETCDBackup object in a final state.
	for _, backup := range backups.Items {
		// Ignore this backup CR if it is not in a final state.
		if backup.Status.Status != backupStateCompleted && backup.Status.Status != backupStateFailed {
			continue
		}
		if newest == nil || backup.Status.FinishedTimestamp.After(newest.Status.FinishedTimestamp.Time) {
			newest = &backup
		}
	}

	if newest != nil {
		for tenantClusterID, instanceStatus := range newest.Status.Instances {
			if instanceStatus.V2.Status == backupStateCompleted {
				ch <- prometheus.MustNewConstMetric(
					creationTimeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V2.CreationTime),
					tenantClusterID,
					"V2",
				)
			}

			if instanceStatus.V3.Status == backupStateCompleted {
				ch <- prometheus.MustNewConstMetric(
					creationTimeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V3.CreationTime),
					tenantClusterID,
					"V3",
				)
			}
		}
	}

	return nil
}

func (d *ETCDBackup) Describe(ch chan<- *prometheus.Desc) error {
	ch <- creationTimeDesc
	ch <- encryptionTimeDesc
	ch <- uploadTimeDesc
	ch <- backupSizeDesc
	ch <- attemptsCounterDesc
	ch <- successCounterDesc
	ch <- failureCounterDesc
	ch <- latestAttemptTimestampDesc
	ch <- latestSuccessTimestampDesc
	return nil
}

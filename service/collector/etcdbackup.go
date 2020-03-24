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
	// This variable holds the name of the lastest CR that has been used to increase the global counters.
	// It is used to increment the counter only once for each new CR.
	lastSent = ""

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

	attemptsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "attempts_count",
			Help:      "Count of attempted backups",
		},
		labels)

	successCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "success_count",
			Help:      "Count of successful backups",
		},
		labels)

	failureCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "failure_count",
			Help:      "Count of failed backups",
		},
		labels)
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

	prometheus.MustRegister(attemptsCounter)
	prometheus.MustRegister(successCounter)
	prometheus.MustRegister(failureCounter)

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

	sendMetricsForVersion := func(tenantClusterID string, status v1alpha1.ETCDInstanceBackupStatus, version string, updateCounter bool) {
		ch <- prometheus.MustNewConstMetric(
			latestAttemptTimestampDesc,
			prometheus.GaugeValue,
			float64(status.FinishedTimestamp.Unix()),
			tenantClusterID,
			version,
		)

		// The updateCounter bool indicates the fact that we need to increment the global counters for this metrics.
		if updateCounter {
			attemptsCounter.WithLabelValues(tenantClusterID, version).Inc()
		}

		if status.Status == backupStateCompleted {
			if updateCounter {
				successCounter.WithLabelValues(tenantClusterID, version).Inc()
			}

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
		} else {
			if updateCounter {
				failureCounter.WithLabelValues(tenantClusterID, version).Inc()
			}
		}
	}

	if newest != nil {
		for _, instanceStatus := range newest.Status.Instances {
			sendMetricsForVersion(instanceStatus.Name, instanceStatus.V2, "v2", newest.Name != lastSent)
			sendMetricsForVersion(instanceStatus.Name, instanceStatus.V3, "v3", newest.Name != lastSent)
		}

		lastSent = newest.Name
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

package collector

import (
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd/metrics"
)

const (
	labelTenantClusterId = "tenant_cluster_id"
)

var (
	namespace                     = "etcd_backup"
	labels                        = []string{labelTenantClusterId}
	constLabels prometheus.Labels = nil

	creationTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "creation_time_ms"),
		"Gauge about the time in ms spent by the ETCD backup creation process.",
		labels,
		constLabels,
	)

	encryptionTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "encryption_time_ms"),
		"Gauge about the time in ms spent by the ETCD backup encryption process.",
		labels,
		constLabels,
	)

	uploadTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "upload_time_ms"),
		"Gauge about the time in ms spent by the ETCD backup upload process.",
		labels,
		constLabels,
	)

	backupSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "size_bytes"),
		"Gauge about the size of the backup file, as seen by S3.",
		labels,
		constLabels,
	)

	attemptsCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "attempts_count"),
		"Count of attempted backups",
		labels,
		constLabels,
	)

	successCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "success_count"),
		"Count of successful backups",
		labels,
		constLabels,
	)

	failureCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "failure_count"),
		"Count of failed backups",
		labels,
		constLabels,
	)

	latestAttemptTimestampDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "latest_attempt"),
		"Timestamp of the latest backup attempt",
		labels,
		constLabels,
	)

	latestSuccessTimestampDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "latest_success"),
		"Timestamp of the latest backup succeeded",
		labels,
		constLabels,
	)
)

type ETCDBackupConfig struct {
	MetricsHolder *metrics.Holder
}

type ETCDBackup struct {
	metricsHolder *metrics.Holder
}

func NewETCDBackup(config ETCDBackupConfig) (*ETCDBackup, error) {
	if config.MetricsHolder == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.MetricsHolder must be defined", config)
	}
	r := &ETCDBackup{
		metricsHolder: config.MetricsHolder,
	}

	return r, nil
}

func (r *ETCDBackup) Collect(ch chan<- prometheus.Metric) error {
	for _, backupMetrics := range r.metricsHolder.GetData() {
		if backupMetrics.CreationTime > 0 {
			ch <- prometheus.MustNewConstMetric(
				creationTimeDesc,
				prometheus.GaugeValue,
				float64(backupMetrics.CreationTime),
				backupMetrics.InstanceName,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			encryptionTimeDesc,
			prometheus.GaugeValue,
			float64(backupMetrics.EncryptionTime),
			backupMetrics.InstanceName,
		)

		if backupMetrics.UploadTime > 0 {
			ch <- prometheus.MustNewConstMetric(
				uploadTimeDesc,
				prometheus.GaugeValue,
				float64(backupMetrics.UploadTime),
				backupMetrics.InstanceName,
			)
		}

		if backupMetrics.BackupFileSize > 0 {
			ch <- prometheus.MustNewConstMetric(
				backupSizeDesc,
				prometheus.GaugeValue,
				float64(backupMetrics.BackupFileSize),
				backupMetrics.InstanceName,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			attemptsCounterDesc,
			prometheus.CounterValue,
			float64(backupMetrics.AttemptsCount),
			backupMetrics.InstanceName,
		)

		ch <- prometheus.MustNewConstMetric(
			successCounterDesc,
			prometheus.CounterValue,
			float64(backupMetrics.SuccessesCount),
			backupMetrics.InstanceName,
		)

		ch <- prometheus.MustNewConstMetric(
			failureCounterDesc,
			prometheus.CounterValue,
			float64(backupMetrics.FailuresCount),
			backupMetrics.InstanceName,
		)

		ch <- prometheus.MustNewConstMetric(
			latestAttemptTimestampDesc,
			prometheus.GaugeValue,
			float64(backupMetrics.LatestAttemptTS),
			backupMetrics.InstanceName,
		)

		if backupMetrics.LatestSuccessTS > 0 {
			ch <- prometheus.MustNewConstMetric(
				latestSuccessTimestampDesc,
				prometheus.GaugeValue,
				float64(backupMetrics.LatestSuccessTS),
				backupMetrics.InstanceName,
			)
		}
	}

	return nil
}

func (r *ETCDBackup) Describe(ch chan<- *prometheus.Desc) error {
	ch <- creationTimeDesc
	ch <- encryptionTimeDesc
	ch <- uploadTimeDesc
	ch <- backupSizeDesc
	ch <- attemptsCounterDesc
	ch <- successCounterDesc
	ch <- failureCounterDesc

	return nil
}

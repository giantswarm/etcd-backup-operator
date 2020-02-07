package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelTenantClusterId  = "tenant_cluster_id"
	metricExpirationHours = 24
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
		"Timestamp of the latest attempted scrape",
		labels,
		constLabels,
	)

	latestSuccessTimestampDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "latest_success"),
		"Timestamp of the latest successful scrape",
		labels,
		constLabels,
	)
)

type ETCDBackupConfig struct {
	ETCDBackupMetrics *ETCDBackupMetrics
}

type ETCDBackup struct {
	ETCDBackupMetrics *ETCDBackupMetrics
}

func NewETCDBackup(config ETCDBackupConfig) (*ETCDBackup, error) {
	r := &ETCDBackup{
		ETCDBackupMetrics: config.ETCDBackupMetrics,
	}

	return r, nil
}

func (r *ETCDBackup) Collect(ch chan<- prometheus.Metric) error {
	r.ETCDBackupMetrics.mux.Lock()
	defer r.ETCDBackupMetrics.mux.Unlock()
	for instance, metrics := range r.ETCDBackupMetrics.data {
		// Check if metric is expired.
		// This is needed to avoid keep sending metrics for deleted tenant clusters.
		diff := time.Now().Sub(metrics.MetricUpdateTS).Hours()
		if diff > metricExpirationHours {
			delete(r.ETCDBackupMetrics.data, instance)
			continue
		}
		if metrics.CreationTime > 0 {
			ch <- prometheus.MustNewConstMetric(
				creationTimeDesc,
				prometheus.GaugeValue,
				float64(metrics.CreationTime),
				instance,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			encryptionTimeDesc,
			prometheus.GaugeValue,
			float64(metrics.EncryptionTime),
			instance,
		)

		if metrics.UploadTime > 0 {
			ch <- prometheus.MustNewConstMetric(
				uploadTimeDesc,
				prometheus.GaugeValue,
				float64(metrics.UploadTime),
				instance,
			)
		}

		if metrics.BackupSize > 0 {
			ch <- prometheus.MustNewConstMetric(
				backupSizeDesc,
				prometheus.GaugeValue,
				float64(metrics.BackupSize),
				instance,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			attemptsCounterDesc,
			prometheus.CounterValue,
			float64(metrics.Attempts),
			instance,
		)

		ch <- prometheus.MustNewConstMetric(
			successCounterDesc,
			prometheus.CounterValue,
			float64(metrics.Successes),
			instance,
		)

		ch <- prometheus.MustNewConstMetric(
			failureCounterDesc,
			prometheus.CounterValue,
			float64(metrics.Failures),
			instance,
		)

		ch <- prometheus.MustNewConstMetric(
			latestAttemptTimestampDesc,
			prometheus.GaugeValue,
			float64(metrics.AttemptTS),
			instance,
		)

		if metrics.SuccessTS > 0 {
			ch <- prometheus.MustNewConstMetric(
				latestSuccessTimestampDesc,
				prometheus.GaugeValue,
				float64(metrics.SuccessTS),
				instance,
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

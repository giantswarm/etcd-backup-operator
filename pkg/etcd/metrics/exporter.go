package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Exporter struct {
}

const (
	labelTenantClusterId = "tenant_cluster_id"
	labelETCDVersion     = "etcd_version"
)

func NewExporter() (*Exporter, error) {
	return &Exporter{}, nil
}

func (h Exporter) Boot(context context.Context) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":8112", nil)
}

var (
	namespace = "etcd_backup"
	labels    = []string{labelTenantClusterId, labelETCDVersion}

	creationTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "creation_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup creation process.",
	}, labels)

	encryptionTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "encryption_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup encryption process.",
	}, labels)

	uploadTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "upload_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup upload process.",
	}, labels)

	backupSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "size_bytes"),
		Help: "Gauge about the size of the backup file, as seen by S3.",
	}, labels)

	attemptsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: prometheus.BuildFQName(namespace, "", "attempts_count"),
		Help: "Count of attempted backups",
	}, labels)

	successCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: prometheus.BuildFQName(namespace, "", "success_count"),
		Help: "Count of successful backups",
	}, labels)

	failureCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: prometheus.BuildFQName(namespace, "", "failure_count"),
		Help: "Count of failed backups",
	}, labels)

	latestAttemptTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "latest_attempt"),
		Help: "Timestamp of the latest backup attempt",
	}, labels)

	latestSuccessTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "latest_success"),
		Help: "Timestamp of the latest backup succeeded",
	}, labels)
)

func (h Exporter) Add(instanceName string, etcdVersion string, metric *BackupAttemptResult) {
	now := time.Now().Unix()

	labels := prometheus.Labels{
		labelTenantClusterId: instanceName,
		labelETCDVersion:     etcdVersion,
	}

	attemptsCounter.With(labels).Inc()
	latestAttemptTimestamp.With(labels).Set(float64(now))

	if metric.Successful {
		successCounter.With(labels).Inc()
		creationTime.With(labels).Set(float64(metric.CreationTimeMeasurement))
		encryptionTime.With(labels).Set(float64(metric.EncryptionTimeMeasurement))
		uploadTime.With(labels).Set(float64(metric.UploadTimeMeasurement))
		backupSize.With(labels).Set(float64(metric.BackupSizeMeasurement))
		latestSuccessTimestamp.With(labels).Set(float64(now))
	} else {
		failureCounter.With(labels).Inc()
	}
}

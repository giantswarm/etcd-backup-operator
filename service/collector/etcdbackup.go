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

	//bk, err := d.g8sClient.BackupV1alpha1().ETCDBackups().Get("testing", v1.GetOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//
	//const longForm = time.RFC3339Nano
	//v2start, _ := time.Parse(longForm, "2020-03-23T16:14:05.747764514Z")
	//v2finish, _ := time.Parse(longForm, "2020-03-23T16:14:19.871264574Z")
	//v3start, _ := time.Parse(longForm, "2020-03-23T16:14:20.011256915Z")
	//v3finish, _ := time.Parse(longForm, "2020-03-23T16:14:36.574252096Z")
	//globalstart, _ := time.Parse(longForm, "2020-03-23T16:14:05.537463552Z")
	//globalfinish, _ := time.Parse(longForm, "2020-03-23T16:14:36.712608237Z")
	//
	//bk.Status = v1alpha1.ETCDBackupStatus{
	//	Instances: map[string]v1alpha1.ETCDInstanceBackupStatusIndex{
	//		"Control Plane": {
	//			Name: "Control Plane",
	//			V2: v1alpha1.ETCDInstanceBackupStatus{
	//				Status: "Completed",
	//				StartedTimestamp: v1alpha1.DeepCopyTime{
	//					Time: v2start,
	//				},
	//				FinishedTimestamp: v1alpha1.DeepCopyTime{
	//					Time: v2finish,
	//				},
	//				LatestError:    "",
	//				CreationTime:   12502,
	//				EncryptionTime: 0,
	//				UploadTime:     1614,
	//				BackupFileSize: 4708407,
	//			},
	//			V3: v1alpha1.ETCDInstanceBackupStatus{
	//				Status: "Completed",
	//				StartedTimestamp: v1alpha1.DeepCopyTime{
	//					Time: v3start,
	//				},
	//				FinishedTimestamp: v1alpha1.DeepCopyTime{
	//					Time: v3finish,
	//				},
	//				LatestError:    "",
	//				CreationTime:   15041,
	//				EncryptionTime: 0,
	//				UploadTime:     1515,
	//				BackupFileSize: 13895102,
	//			},
	//		},
	//	},
	//	Status: "Completed",
	//	StartedTimestamp: v1alpha1.DeepCopyTime{
	//		Time: globalstart,
	//	},
	//	FinishedTimestamp: v1alpha1.DeepCopyTime{
	//		Time: globalfinish,
	//	},
	//}
	//
	//d.g8sClient.BackupV1alpha1().ETCDBackups().UpdateStatus(bk)

	//attemptsCounter.WithLabelValues("test", "V2").Inc()

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
			ch <- prometheus.MustNewConstMetric(
				latestAttemptTimestampDesc,
				prometheus.GaugeValue,
				float64(instanceStatus.V2.FinishedTimestamp.Unix()),
				tenantClusterID,
				"V2",
			)

			ch <- prometheus.MustNewConstMetric(
				latestAttemptTimestampDesc,
				prometheus.GaugeValue,
				float64(instanceStatus.V3.FinishedTimestamp.Unix()),
				tenantClusterID,
				"V3",
			)

			if instanceStatus.V2.Status == backupStateCompleted {
				ch <- prometheus.MustNewConstMetric(
					creationTimeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V2.CreationTime),
					tenantClusterID,
					"V2",
				)

				ch <- prometheus.MustNewConstMetric(
					encryptionTimeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V2.EncryptionTime),
					tenantClusterID,
					"V2",
				)

				ch <- prometheus.MustNewConstMetric(
					uploadTimeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V2.UploadTime),
					tenantClusterID,
					"V2",
				)

				ch <- prometheus.MustNewConstMetric(
					backupSizeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V2.BackupFileSize),
					tenantClusterID,
					"V2",
				)

				ch <- prometheus.MustNewConstMetric(
					latestSuccessTimestampDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V2.FinishedTimestamp.Unix()),
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

				ch <- prometheus.MustNewConstMetric(
					encryptionTimeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V3.EncryptionTime),
					tenantClusterID,
					"V3",
				)

				ch <- prometheus.MustNewConstMetric(
					uploadTimeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V3.UploadTime),
					tenantClusterID,
					"V3",
				)

				ch <- prometheus.MustNewConstMetric(
					backupSizeDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V3.BackupFileSize),
					tenantClusterID,
					"V3",
				)

				ch <- prometheus.MustNewConstMetric(
					latestSuccessTimestampDesc,
					prometheus.GaugeValue,
					float64(instanceStatus.V3.FinishedTimestamp.Unix()),
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
	ch <- latestAttemptTimestampDesc
	ch <- latestSuccessTimestampDesc
	return nil
}

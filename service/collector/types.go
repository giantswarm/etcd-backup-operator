package collector

import (
	"github.com/giantswarm/etcd-backup/metrics"
	"time"
)

type singleEtcdBackupMetric struct {
	Attempts       int32
	Successes      int32
	Failures       int32
	CreationTime   int64
	EncryptionTime int64
	UploadTime     int64
	BackupSize     int64
	AttemptTS      int64
	SuccessTS      int64
}

type ETCDBackupMetrics struct {
	data map[string]*singleEtcdBackupMetric
}

func (m *ETCDBackupMetrics) Update(instanceName string, metrics *metrics.BackupMetrics) {
	if m.data == nil {
		m.data = make(map[string]*singleEtcdBackupMetric)
	}
	current := m.data[instanceName]
	if current == nil {
		current = &singleEtcdBackupMetric{}
	}

	now := time.Now().Unix()

	current.AttemptTS = now
	current.Attempts = current.Attempts + 1
	if metrics.Successful {
		current.Successes = current.Successes + 1
		current.CreationTime = metrics.CreationTimeMeasurement
		current.EncryptionTime = metrics.EncryptionTimeMeasurement
		current.UploadTime = metrics.UploadTimeMeasurement
		current.BackupSize = metrics.BackupSizeMeasurement
		current.SuccessTS = now
	} else {
		current.Failures = current.Failures + 1
	}

	m.data[instanceName] = current
}

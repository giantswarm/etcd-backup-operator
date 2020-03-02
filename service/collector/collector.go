package collector

import (
	"sync"
	"time"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd/metrics"
)

type singleEtcdBackupMetric struct {
	Attempts       int32
	AttemptTS      int64
	BackupSize     int64
	CreationTime   int64
	EncryptionTime int64
	Failures       int32
	Successes      int32
	SuccessTS      int64
	UploadTime     int64

	MetricUpdateTS time.Time
}

type ETCDBackupMetrics struct {
	data map[string]*singleEtcdBackupMetric
	mux  sync.Mutex
}

func (m *ETCDBackupMetrics) Update(instanceName string, metrics *metrics.BackupMetrics) {
	m.mux.Lock()
	defer m.mux.Unlock()
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

	current.MetricUpdateTS = time.Now()

	m.data[instanceName] = current
}

package metrics

import (
	"sync"
	"time"
)

type Holder struct {
	data map[string]*instanceBackupMetrics
	mux  sync.Mutex
}

func NewMetricsHolder() (*Holder, error) {
	return &Holder{}, nil
}

// Returns a copy of the current metrics data in a synchronized way.
func (h Holder) GetData() []instanceBackupMetrics {
	h.mux.Lock()
	defer h.mux.Unlock()

	var ret []instanceBackupMetrics

	for _, metrics := range h.data {
		ret = append(ret, *metrics)
	}

	return ret
}

func (h Holder) Add(instanceName string, metric *BackupAttemptResult) {
	h.mux.Lock()
	defer h.mux.Unlock()

	if h.data == nil {
		h.data = make(map[string]*instanceBackupMetrics)
	}

	current := h.data[instanceName]
	if current == nil {
		current = &instanceBackupMetrics{}
	}

	now := time.Now().Unix()
	current.LatestAttemptTS = now
	current.AttemptsCount = current.AttemptsCount + 1
	if metric.Successful {
		current.SuccessesCount = current.SuccessesCount + 1
		current.CreationTime = metric.CreationTimeMeasurement
		current.EncryptionTime = metric.EncryptionTimeMeasurement
		current.UploadTime = metric.UploadTimeMeasurement
		current.BackupFileSize = metric.BackupSizeMeasurement
		current.LatestSuccessTS = now
	} else {
		current.FailuresCount = current.FailuresCount + 1
	}

	h.data[instanceName] = current
}

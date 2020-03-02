package metrics

import (
	"sync"
)

type Holder struct {
	data map[string]instanceBackupMetrics
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
		ret = append(ret, metrics)
	}

	return ret
}

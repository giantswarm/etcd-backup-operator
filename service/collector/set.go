package collector

import (
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd/metrics"
)

type SetConfig struct {
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	MetricsHolder *metrics.Holder
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the initialization and prevents some weird import mess so we do not
// have to alias packages. There is also the benefit of the helper type kept
// private so we do not need to expose this magic.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must be defined", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must be defined", config)
	}
	if config.MetricsHolder == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.MetricsHolder must be defined", config)
	}

	etcdBackup, err := NewETCDBackup(ETCDBackupConfig{
		MetricsHolder: config.MetricsHolder,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				etcdBackup,
			},
			Logger: config.Logger,
		}

		collectorSet, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Set{
		Set: collectorSet,
	}

	return s, nil
}

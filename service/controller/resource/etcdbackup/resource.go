package etcdbackup

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

const (
	Name = "todo"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	logger       micrologger.Logger
	k8sClient    k8sclient.Interface
	stateMachine state.Machine
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.k8sClient must not be empty", config)
	}

	r := &Resource{
		logger:    config.Logger,
		k8sClient: config.K8sClient,
	}

	r.configureStateMachine()

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

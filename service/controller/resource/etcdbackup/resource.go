package etcdbackup

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/etcd-backup-operator/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

const (
	Name = "todo"
)

type Config struct {
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	ETCDv2Settings giantnetes.ETCDv2Settings
	ETCDv3Settings giantnetes.ETCDv3Settings
}

type Resource struct {
	logger         micrologger.Logger
	k8sClient      k8sclient.Interface
	stateMachine   state.Machine
	etcdV2Settings giantnetes.ETCDv2Settings
	etcdV3Settings giantnetes.ETCDv3Settings
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.k8sClient must not be empty", config)
	}
	if !config.ETCDv2Settings.AreComplete() && !config.ETCDv3Settings.AreComplete() {
		return nil, microerror.Maskf(invalidConfigError, "Either %T.ETCDv2Settings or %T.ETCDv3Settings must be defined", config, config)
	}

	r := &Resource{
		logger:         config.Logger,
		k8sClient:      config.K8sClient,
		etcdV2Settings: config.ETCDv2Settings,
		etcdV3Settings: config.ETCDv3Settings,
	}

	r.configureStateMachine()

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

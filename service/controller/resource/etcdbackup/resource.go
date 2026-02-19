package etcdbackup

import (
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/etcd-backup-operator/v4/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/v4/pkg/storage"
	"github.com/giantswarm/etcd-backup-operator/v4/service/controller/resource/etcdbackup/internal/state"
)

const (
	Name = "etcd-backup"
)

type Config struct {
	K8sClient                   k8sclient.Interface
	Logger                      micrologger.Logger
	ETCDv3Settings              giantnetes.ETCDv3Settings
	EncryptionPwd               string
	Installation                string
	Uploader                    storage.Uploader
	SkipManagementClusterBackup bool
}

type Resource struct {
	logger       micrologger.Logger
	k8sClient    k8sclient.Interface
	stateMachine state.Machine

	etcdV3Settings              giantnetes.ETCDv3Settings
	encryptionPwd               string
	installation                string
	uploader                    storage.Uploader
	skipManagementClusterBackup bool
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.k8sClient must not be empty", config)
	}
	if !config.SkipManagementClusterBackup && !config.ETCDv3Settings.AreComplete() {
		return nil, microerror.Maskf(invalidConfigError, "%T.ETCDv3Settings must be defined", config)
	}
	if config.Installation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installation must not be empty", config)
	}
	if config.Uploader == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Uploader must not be empty", config)
	}

	r := &Resource{
		logger:                      config.Logger,
		k8sClient:                   config.K8sClient,
		etcdV3Settings:              config.ETCDv3Settings,
		encryptionPwd:               config.EncryptionPwd,
		installation:                config.Installation,
		uploader:                    config.Uploader,
		skipManagementClusterBackup: config.SkipManagementClusterBackup,
	}

	r.configureStateMachine()

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

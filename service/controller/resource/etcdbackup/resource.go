package etcdbackup

import (
	"github.com/giantswarm/etcd-backup-operator/service/collector"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/storage"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "etcdbackup"
)

type Config struct {
	K8sClient         k8sclient.Interface
	Logger            micrologger.Logger
	S3Config          storage.S3Uploader
	ETCDv2Settings    resource.ETCDv2Settings
	ETCDv3Settings    resource.ETCDv3Settings
	ETCDBackupMetrics *collector.ETCDBackupMetrics
}

type Resource struct {
	logger            micrologger.Logger
	K8sClient         k8sclient.Interface
	S3Config          storage.S3Uploader
	ETCDv2Settings    resource.ETCDv2Settings
	ETCDv3Settings    resource.ETCDv3Settings
	ETCDBackupMetrics *collector.ETCDBackupMetrics
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger:            config.Logger,
		K8sClient:         config.K8sClient,
		S3Config:          config.S3Config,
		ETCDv2Settings:    config.ETCDv2Settings,
		ETCDv3Settings:    config.ETCDv3Settings,
		ETCDBackupMetrics: config.ETCDBackupMetrics,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

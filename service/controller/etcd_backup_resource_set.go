package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"

	"github.com/giantswarm/etcd-backup-operator/v2/service/controller/resource/etcdbackup"
)

func validateETCDBackupResourceSetConfigConfig(config ETCDBackupConfig) error {
	if !config.ETCDv2Settings.AreComplete() && !config.ETCDv3Settings.AreComplete() {
		return microerror.Maskf(invalidConfigError, "Either %T.ETCDv2Settings or %T.ETCDv3Settings must be defined", config, config)
	}
	if config.Uploader == nil {
		return microerror.Maskf(invalidConfigError, "%T.Uploader must be defined", config)
	}
	return nil
}

func newETCDBackupResourceSet(config ETCDBackupConfig) ([]resource.Interface, error) {
	var err error
	err = validateETCDBackupResourceSetConfigConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var etcdBackupResource resource.Interface
	{
		c := etcdbackup.Config{
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ETCDv2Settings: config.ETCDv2Settings,
			ETCDv3Settings: config.ETCDv3Settings,
			EncryptionPwd:  config.EncryptionPwd,
			Uploader:       config.Uploader,
		}

		etcdBackupResource, err = etcdbackup.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		etcdBackupResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}

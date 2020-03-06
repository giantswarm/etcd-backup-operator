package controller

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd/metrics"
	"github.com/giantswarm/etcd-backup-operator/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/pkg/storage"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup"
)

type etcdBackupResourceSetConfig struct {
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	ETCDv2Settings giantnetes.ETCDv2Settings
	ETCDv3Settings giantnetes.ETCDv3Settings
	EncryptionPwd  string
	MetricsHolder  *metrics.Exporter
	Uploader       storage.Uploader
}

func validateETCDBackupResourceSetConfigConfig(config etcdBackupResourceSetConfig) error {
	if !config.ETCDv2Settings.AreComplete() && !config.ETCDv3Settings.AreComplete() {
		return microerror.Maskf(invalidConfigError, "Either %T.ETCDv2Settings or %T.ETCDv3Settings must be defined", config, config)
	}
	if config.MetricsHolder == nil {
		return microerror.Maskf(invalidConfigError, "%T.MetricsHolder must be defined", config)
	}
	if config.Uploader == nil {
		return microerror.Maskf(invalidConfigError, "%T.Uploader must be defined", config)
	}
	return nil
}

func newETCDBackupResourceSet(config etcdBackupResourceSetConfig) (*controller.ResourceSet, error) {
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
			MetricsHolder:  config.MetricsHolder,
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

	// handlesFunc defines which objects you want to get into your controller, e.g. which objects you want to watch.
	handlesFunc := func(obj interface{}) bool {
		// ETCDBackup: By default this will handle all objects of the type your controller is watching.
		// Your controller is watching a certain kubernetes type, so why do we need to check again?
		// Because there might be a change in the object structure - e.g. the type `AWSConfig` object might have the field
		// availabilityZones recently, but older ones don't, and you don't want to handle those.
		//
		// Normally we use this to filter objects containing the expected `versionbundle` version.
		// So two versions of your operator don't accidentally reconcile the same CR.
		return true
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

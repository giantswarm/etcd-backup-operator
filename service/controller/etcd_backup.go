package controller

import (
	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/etcd-backup-operator/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/pkg/project"
	"github.com/giantswarm/etcd-backup-operator/pkg/storage"
)

type ETCDBackupConfig struct {
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	ETCDv2Settings giantnetes.ETCDv2Settings
	ETCDv3Settings giantnetes.ETCDv3Settings
	EncryptionPwd  string
	Uploader       storage.Uploader
}

type ETCDBackup struct {
	*controller.Controller
}

func validateETCDBackupConfig(config ETCDBackupConfig) error {
	if !config.ETCDv2Settings.AreComplete() && !config.ETCDv3Settings.AreComplete() {
		return microerror.Maskf(invalidConfigError, "Either %T.ETCDv2Settings or %T.ETCDv3Settings must be defined", config, config)
	}
	if config.Uploader == nil {
		return microerror.Maskf(invalidConfigError, "%T.Uploader must be defined", config)
	}
	return nil
}

func NewETCDBackup(config ETCDBackupConfig) (*ETCDBackup, error) {
	var err error
	err = validateETCDBackupConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	resourceSets, err := newETCDBackupResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(backupv1alpha1.ETCDBackup)
			},
			Name: project.Name() + "-etcd-backup-controller",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &ETCDBackup{
		Controller: operatorkitController,
	}

	return c, nil
}

func newETCDBackupResourceSets(config ETCDBackupConfig) ([]*controller.ResourceSet, error) {
	var err error
	err = validateETCDBackupConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var resourceSet *controller.ResourceSet
	{
		c := etcdBackupResourceSetConfig{
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ETCDv2Settings: config.ETCDv2Settings,
			ETCDv3Settings: config.ETCDv3Settings,
			EncryptionPwd:  config.EncryptionPwd,
			Uploader:       config.Uploader,
		}

		resourceSet, err = newETCDBackupResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}

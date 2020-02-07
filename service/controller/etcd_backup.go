package controller

import (
	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/etcd-backup-operator/pkg/project"
)

type ETCDBackupConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type EtcdBackup struct {
	*controller.Controller
}

func validateETCDBackupConfig(config ETCDBackupConfig) error {
	return nil
}

func NewETCDBackup(config ETCDBackupConfig) (*EtcdBackup, error) {
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
			CRD:          backupv1alpha1.NewEtcdBackupCRD(),
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

	c := &EtcdBackup{
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
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
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

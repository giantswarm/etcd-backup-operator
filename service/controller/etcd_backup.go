package controller

import (
	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/etcd-backup-operator/service/collector"
	etcdresource "github.com/giantswarm/etcd-backup-operator/service/controller/resource"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/storage"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/etcd-backup-operator/pkg/project"
)

type ETCDBackupConfig struct {
	K8sClient         k8sclient.Interface
	Logger            micrologger.Logger
	S3Config          storage.S3Uploader
	ETCDv2Settings    etcdresource.ETCDv2Settings
	ETCDv3Settings    etcdresource.ETCDv3Settings
	ETCDBackupMetrics *collector.ETCDBackupMetrics
}

type EtcdBackup struct {
	*controller.Controller
}

func NewETCDBackup(config ETCDBackupConfig) (*EtcdBackup, error) {
	var err error

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

	var resourceSet *controller.ResourceSet
	{
		c := etcdBackupResourceSetConfig{
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			S3Config:          config.S3Config,
			ETCDv2Settings:    config.ETCDv2Settings,
			ETCDv3Settings:    config.ETCDv3Settings,
			ETCDBackupMetrics: config.ETCDBackupMetrics,
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

// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/storage"
	"sync"

	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/etcd-backup-operator/flag"
	"github.com/giantswarm/etcd-backup-operator/pkg/project"
	"github.com/giantswarm/etcd-backup-operator/service/collector"
	"github.com/giantswarm/etcd-backup-operator/service/controller"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	Version *version.Service

	bootOnce             sync.Once
	etcdBackupController *controller.EtcdBackup
	operatorCollector    *collector.Set
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	var serviceAddress string
	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}
	if config.Flag.Service.Kubernetes.KubeConfig == "" {
		serviceAddress = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
	} else {
		serviceAddress = ""
	}
	if config.Viper.GetString(config.Flag.Service.S3.Bucket) == "" {
		return nil, microerror.Maskf(invalidConfigError, "S3Uploader bucket must not be empty.")
	}
	if config.Viper.GetString(config.Flag.Service.S3.Region) == "" {
		return nil, microerror.Maskf(invalidConfigError, "S3Uploader region must not be empty.")
	}
	// If ETCDv2 data dir is empty, all the ETCDv3 settings must be specified
	if config.Viper.GetString(config.Flag.Service.ETCDv2.DataDir) == "" {
		if config.Viper.GetString(config.Flag.Service.ETCDv3.Endpoints) == "" ||
			config.Viper.GetString(config.Flag.Service.ETCDv3.Key) == "" ||
			config.Viper.GetString(config.Flag.Service.ETCDv3.CaCert) == "" ||
			config.Viper.GetString(config.Flag.Service.ETCDv3.Cert) == "" {
			return nil, microerror.Maskf(invalidConfigError, "One of ETCDv2 or ETCDv3 settings must be specified.")
		}
	}
	// If any of the ETCDv3 Flags are set, than all have to be set.
	if config.Viper.GetString(config.Flag.Service.ETCDv3.Endpoints) != "" ||
		config.Viper.GetString(config.Flag.Service.ETCDv3.Key) != "" ||
		config.Viper.GetString(config.Flag.Service.ETCDv3.CaCert) != "" ||
		config.Viper.GetString(config.Flag.Service.ETCDv3.Cert) != "" {
		if config.Viper.GetString(config.Flag.Service.ETCDv3.Endpoints) == "" ||
			config.Viper.GetString(config.Flag.Service.ETCDv3.Key) == "" ||
			config.Viper.GetString(config.Flag.Service.ETCDv3.CaCert) == "" ||
			config.Viper.GetString(config.Flag.Service.ETCDv3.Cert) == "" {
			return nil, microerror.Maskf(invalidConfigError, "Endpoints, Key, CaCert and Cert keys are all required if one of them is set.")
		}
	}

	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:    serviceAddress,
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger: config.Logger,
			SchemeBuilder: k8sclient.SchemeBuilder{
				backupv1alpha1.AddToScheme,
			},
			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	etcdBackupMetrics := collector.ETCDBackupMetrics{}

	uploader := storage.NewS3Uploader(
		config.Viper.GetString(config.Flag.Service.S3.Bucket),
		config.Viper.GetString(config.Flag.Service.S3.Region),
	)

	var etcdBackupController *controller.EtcdBackup
	{
		c := controller.ETCDBackupConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,
			ETCDv2Settings: resource.ETCDv2Settings{
				DataDir: config.Viper.GetString(config.Flag.Service.ETCDv2.DataDir),
			},
			ETCDv3Settings: resource.ETCDv3Settings{
				Endpoints: config.Viper.GetString(config.Flag.Service.ETCDv3.Endpoints),
				CaCert:    config.Viper.GetString(config.Flag.Service.ETCDv3.CaCert),
				Key:       config.Viper.GetString(config.Flag.Service.ETCDv3.Key),
				Cert:      config.Viper.GetString(config.Flag.Service.ETCDv3.Cert),
			},
			ETCDBackupMetrics: &etcdBackupMetrics,
			Uploader:          uploader,
		}

		etcdBackupController, err = controller.NewETCDBackup(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			K8sClient:         k8sClient.K8sClient(),
			Logger:            config.Logger,
			ETCDBackupMetrics: &etcdBackupMetrics,
		}

		operatorCollector, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.Config{
			Description:    project.Description(),
			GitCommit:      project.GitSHA(),
			Name:           project.Name(),
			Source:         project.Source(),
			Version:        project.Version(),
			VersionBundles: []versionbundle.Bundle{project.NewVersionBundle()},
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:             sync.Once{},
		etcdBackupController: etcdBackupController,
		operatorCollector:    operatorCollector,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx)

		go s.etcdBackupController.Boot(ctx)
	})
}

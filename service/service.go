// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"crypto/tls"
	"os"
	"sync"

	backupv1alpha1 "github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v7/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/etcd-backup-operator/v3/flag"
	"github.com/giantswarm/etcd-backup-operator/v3/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/v3/pkg/project"
	"github.com/giantswarm/etcd-backup-operator/v3/pkg/storage"
	"github.com/giantswarm/etcd-backup-operator/v3/service/collector"
	"github.com/giantswarm/etcd-backup-operator/v3/service/controller"
	"github.com/giantswarm/etcd-backup-operator/v3/service/controller/key"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	logger  micrologger.Logger
	version *version.Service

	bootOnce             sync.Once
	etcdBackupController *controller.ETCDBackup
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
	// If ETCDv2 data dir is empty, all the ETCDv3 settings must be specified.
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
				infrastructurev1alpha3.AddToScheme,
				providerv1alpha1.AddToScheme,
				capi.AddToScheme,
			},
			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var etcdBackupController *controller.ETCDBackup
	{
		uploader, err := storage.NewS3Upload(storage.S3Config{
			AccessKeyID:     os.Getenv(key.EnvAWSAccessKeyID),
			Bucket:          config.Viper.GetString(config.Flag.Service.S3.Bucket),
			Region:          config.Viper.GetString(config.Flag.Service.S3.Region),
			SecretAccessKey: os.Getenv(key.EnvAWSSecretAccessKey),
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		skipMCBackup := true // config.Viper.GetBool(config.Flag.Service.SkipManagementClusterBackup)
		// fmt.Printf("\n\nloaded viper config skipMCBackup %t\n\n", skipMCBackup)

		var tlsConfig *tls.Config = nil
		/*if !skipMCBackup {
			tlsConfig, err = key.TLSConfigFromCertFiles(
				config.Viper.GetString(config.Flag.Service.ETCDv3.CaCert),
				config.Viper.GetString(config.Flag.Service.ETCDv3.Cert),
				config.Viper.GetString(config.Flag.Service.ETCDv3.Key),
			)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}*/

		c := controller.ETCDBackupConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,
			ETCDv2Settings: giantnetes.ETCDv2Settings{
				DataDir: config.Viper.GetString(config.Flag.Service.ETCDv2.DataDir),
			},
			ETCDv3Settings: giantnetes.ETCDv3Settings{
				Endpoints: config.Viper.GetString(config.Flag.Service.ETCDv3.Endpoints),
				TLSConfig: tlsConfig,
			},
			EncryptionPwd:               os.Getenv(key.EncryptionPassword),
			Installation:                config.Viper.GetString(config.Flag.Service.Installation),
			SentryDSN:                   config.Viper.GetString(config.Flag.Service.Sentry.DSN),
			SkipManagementClusterBackup: skipMCBackup,
			Uploader:                    uploader,
		}

		etcdBackupController, err = controller.NewETCDBackup(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,
		}

		operatorCollector, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.Config{
			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		logger:  config.Logger,
		version: versionService,

		bootOnce:             sync.Once{},
		etcdBackupController: etcdBackupController,
		operatorCollector:    operatorCollector,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go func() {
			err := s.operatorCollector.Boot(ctx)
			if err != nil {
				s.logger.LogCtx(ctx, "level", "error", "message", "failed to boot collector", "stack", microerror.JSON(err))
				os.Exit(1)
			}

		}()
		go s.etcdBackupController.Boot(ctx)
	})
}

func (s *Service) GetVersion() *version.Service {
	return s.version
}

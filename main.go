package main

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/giantswarm/etcd-backup-operator/v5/flag"
	"github.com/giantswarm/etcd-backup-operator/v5/pkg/project"
	"github.com/giantswarm/etcd-backup-operator/v5/server"
	"github.com/giantswarm/etcd-backup-operator/v5/service"
)

var (
	f *flag.Flag = flag.New()
)

func main() {
	err := mainE(context.Background())
	if err != nil {
		panic(microerror.JSON(err))
	}
}

func mainE(ctx context.Context) error {
	var err error

	// Initialize controller-runtime logger to suppress the
	// "log.SetLogger(...) was never called" warning. All operational
	// logging goes through micrologger, so we discard controller-runtime logs.
	ctrl.SetLogger(logr.Discard())

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	serverFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			c := service.Config{
				Logger: logger,

				Flag:  f,
				Viper: v,
			}

			newService, err = service.New(c)
			if err != nil {
				panic(microerror.JSON(err))
			}

			go newService.Boot(ctx)
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  logger,
				Service: newService,

				Viper: v,
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(microerror.JSON(err))
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        logger,
			ServerFactory: serverFactory,

			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.KubeConfig, "", "KubeConfig (formatted as JSON string) used to connect to Kubernetes. When empty other settings are used.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().Bool(f.Service.SkipManagementClusterBackup, false, "Skip management cluster backup.")
	daemonCommand.PersistentFlags().String(f.Service.BackupDestination, "", "Backup destination is a filter for the ETCDBackup CRs. This is useful when running multiple instances of the operator in the same cluster.")
	daemonCommand.PersistentFlags().String(f.Service.S3.Bucket, "", "AWS S3 Bucket name.")
	daemonCommand.PersistentFlags().String(f.Service.S3.Region, "", "AWS S3 Region name.")
	daemonCommand.PersistentFlags().String(f.Service.S3.Endpoint, "", "Custom AWS S3 Endpoint.")
	daemonCommand.PersistentFlags().Bool(f.Service.S3.ForcePathStyle, false, "Enable path-style S3 URLs.")
	daemonCommand.PersistentFlags().String(f.Service.ETCDv3.Cert, "", "Client certificate for ETCD v3 connection")
	daemonCommand.PersistentFlags().String(f.Service.ETCDv3.CaCert, "", "Client CA certificate for ETCD v3 connection")
	daemonCommand.PersistentFlags().String(f.Service.ETCDv3.Key, "", "Client private key for ETCD v3 connection")
	daemonCommand.PersistentFlags().String(f.Service.ETCDv3.Endpoints, "", "Endpoints for ETCD v3 connection")
	daemonCommand.PersistentFlags().String(f.Service.Installation, "", "Name of the installation")
	daemonCommand.PersistentFlags().String(f.Service.Sentry.DSN, "", "DSN of the Sentry instance to forward errors to.")
	daemonCommand.PersistentFlags().Bool(f.Service.EnableIRSA, false, "Enable IAM Roles for Service Accounts (IRSA) for S3 access.")

	err = newCommand.CobraCommand().Execute()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

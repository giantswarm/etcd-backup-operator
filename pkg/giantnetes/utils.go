package giantnetes

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	"github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	certificateLabel      = "giantswarm.io/certificate"
	certificateLabelValue = "calico-etcd-client"
)

type Utils struct {
	logger    micrologger.Logger
	K8sClient k8sclient.Interface
}

type Cluster struct {
	clusterID        string
	clusterNamespace string
	provider         string
}

func NewUtils(logger micrologger.Logger, client k8sclient.Interface) (*Utils, error) {
	if logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if client == nil {
		return nil, microerror.Maskf(invalidConfigError, "client must not be empty")
	}

	return &Utils{
		logger:    logger,
		K8sClient: client,
	}, nil
}

func (u *Utils) GetTenantClusters(ctx context.Context, backup v1alpha1.ETCDBackup) ([]ETCDInstance, error) {
	var instances []ETCDInstance

	clusterList, err := u.getAllGuestClusters(ctx, u.K8sClient.CtrlClient())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	u.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("Found %d tenant clusters", len(clusterList)))

	for _, cluster := range clusterList {
		u.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("Preparing instance entry for tenant clusters %s", cluster.clusterID))

		// Check if the cluster release version has support for ETCD backup.
		versionSupported, err := u.checkClusterVersionSupport(ctx, cluster)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to check release version for cluster %s", cluster.clusterID), "reason", err)
			continue
		}
		if !versionSupported {
			u.logger.LogCtx(ctx, "level", "warning", "msg", fmt.Sprintf("Cluster %s is too old for etcd backup. Skipping.", cluster.clusterID))
			continue
		}

		// Fetch ETCD certs.
		certs, err := u.getEtcdTLSCfg(ctx, cluster.clusterID, cluster.clusterNamespace)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to fetch etcd certs for cluster %s", cluster.clusterID), "reason", err)
			continue
		}
		// Write ETCD certs to tmpdir.
		err = u.createCertFiles(cluster.clusterID, certs)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to write etcd certs to tmpdir for cluster %s", cluster.clusterID), "reason", err)
			continue
		}

		// Fetch ETCD endpoint.
		etcdEndpoint, err := u.getEtcdEndpoint(ctx, cluster)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to fetch etcd endpoint for cluster %s", cluster.clusterID), "reason", err)
			continue
		}

		instances = append(instances, ETCDInstance{
			Name:   cluster.clusterID,
			ETCDv2: ETCDv2Settings{},
			ETCDv3: ETCDv3Settings{
				Endpoints: etcdEndpoint,
				CaCert:    certs.CAFile,
				Cert:      certs.CrtFile,
				Key:       certs.KeyFile,
			},
		})
	}
	return instances, nil
}

// Check if cluster release version has guest cluster backup support.
func (u *Utils) checkClusterVersionSupport(ctx context.Context, cluster Cluster) (bool, error) {
	crdClient := u.K8sClient.CtrlClient()

	switch cluster.provider {
	case awsCAPI:
		{
			// Cluster API AWS backups are always supported.
			return true, nil
		}
	case azure:
		{
			crd := providerv1alpha1.AzureConfig{}
			err := crdClient.Get(ctx, client.ObjectKey{Namespace: cluster.clusterNamespace, Name: cluster.clusterID}, &crd)
			if err != nil {
				return false, microerror.Maskf(executionFailedError, fmt.Sprintf("failed to get azure crd %#q with error %#q", cluster.clusterID, err))
			}
			var version string
			{
				version = crd.Spec.VersionBundle.Version
				if version == "" {
					// CAPI clusters still have an AzureConfig, but they don't have the Spec.VersionBundle.Version field set.
					// They save the version in a label.
					version = crd.Labels[label.ReleaseVersion]
				}
			}
			if version == "" {
				return false, microerror.Maskf(executionFailedError, fmt.Sprintf("failed to get cluster version from AzureConfig %#q", cluster.clusterID))
			}
			return stringVersionCmp(version, semver.New("0.0.0"), azureSupportFrom)
		}
	case kvm:
		{
			// KVM backups are always supported.
			return true, nil
		}
	}
	return false, nil
}

// Fetch ETCD client certs.
func (u *Utils) getEtcdTLSCfg(ctx context.Context, clusterID string, clusterNamespace string) (*TLSClientConfig, error) {
	k8sClient := u.K8sClient.CtrlClient()
	secrets := v1.SecretList{}
	err := k8sClient.List(ctx, &secrets, client.MatchingLabels{
		label.Cluster:    clusterID,
		certificateLabel: certificateLabelValue,
	})
	if err != nil {
		return nil, microerror.Maskf(executionFailedError, "error getting etcd client certificates for guest cluster %#q with error %#q", clusterID, err)
	}

	if len(secrets.Items) != 1 {
		return nil, microerror.Maskf(executionFailedError, "expected exactly 1 secret with %s=%q and %s=%q, got %d", label.Cluster, clusterID, certificateLabel, certificateLabelValue, len(secrets.Items))
	}

	secret := secrets.Items[0]

	certs := &TLSClientConfig{
		CAData:  secret.Data["ca"],
		KeyData: secret.Data["key"],
		CrtData: secret.Data["crt"],
	}

	return certs, nil
}

// Fetch guest cluster ETCD endpoint.
func (u *Utils) getEtcdEndpoint(ctx context.Context, cluster Cluster) (string, error) {
	var etcdEndpoint string
	crdClient := u.K8sClient.CtrlClient()

	switch cluster.provider {
	case awsCAPI:
		{
			crd := v1alpha3.AWSCluster{}
			err := crdClient.Get(ctx, client.ObjectKey{Name: cluster.clusterID, Namespace: cluster.clusterNamespace}, &crd)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error getting aws crd for guest cluster %#q with error %#q", cluster.clusterID, err)
			}
			etcdEndpoint = AwsCAPIEtcdEndpoint(cluster.clusterID, crd.Spec.Cluster.DNS.Domain)
			break
		}
	case azure:
		{
			crd := providerv1alpha1.AzureConfig{}
			err := crdClient.Get(ctx, client.ObjectKey{Name: cluster.clusterID, Namespace: cluster.clusterNamespace}, &crd)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error getting azure crd for guest cluster %#q with error %#q", cluster.clusterID, err)
			}
			etcdEndpoint = AzureEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	case kvm:
		{
			crd := providerv1alpha1.KVMConfig{}
			err := crdClient.Get(ctx, client.ObjectKey{Name: cluster.clusterID, Namespace: cluster.clusterNamespace}, &crd)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error getting kvm crd for guest cluster %#q with error %#q", cluster.clusterID, err)
			}
			etcdEndpoint = KVMEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	}

	// We already check for unknown provider at the start.
	return etcdEndpoint, nil
}

// Create cert files in tmp dir from certConfig and saves filenames back.
func (u *Utils) createCertFiles(clusterID string, certConfig *TLSClientConfig) error {
	tmpDir, err := ioutil.TempDir("", clusterID)
	if err != nil {
		return microerror.Mask(err)
	}

	// cert
	err = ioutil.WriteFile(CertFile(clusterID, tmpDir), certConfig.CrtData, fileMode)
	if err != nil {
		return microerror.Maskf(executionFailedError, "failed to write crt file %#q with error %#q", CertFile(clusterID, tmpDir), err)
	}
	certConfig.CrtFile = CertFile(clusterID, tmpDir)

	// key
	err = ioutil.WriteFile(KeyFile(clusterID, tmpDir), certConfig.KeyData, fileMode)
	if err != nil {
		return microerror.Maskf(executionFailedError, "failed to write key file %#q with error %#q", KeyFile(clusterID, tmpDir), err)
	}
	certConfig.KeyFile = KeyFile(clusterID, tmpDir)

	// ca
	err = ioutil.WriteFile(CAFile(clusterID, tmpDir), certConfig.CAData, fileMode)
	if err != nil {
		return microerror.Maskf(executionFailedError, "failed to write ca file %#q with error %#q", CAFile(clusterID, tmpDir), err)
	}
	certConfig.CAFile = CAFile(clusterID, tmpDir)

	return nil
}

// Fetch all guest clusters IDs in host cluster.
func (u *Utils) getAllGuestClusters(ctx context.Context, crdCLient client.Client) ([]Cluster, error) {
	var clusterList []Cluster
	anySuccess := false

	// AWS Cluster API
	{
		crdList := v1alpha3.AWSClusterList{}
		err := crdCLient.List(ctx, &crdList)
		if err == nil {
			anySuccess = true
			for _, awsClusterObj := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if awsClusterObj.DeletionTimestamp == nil {
					clusterList = append(clusterList, Cluster{awsClusterObj.Name, awsClusterObj.Namespace, awsCAPI})
				}
			}
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing AWSClusters: %s", err))
		}
	}

	// Azure
	{
		crdList := providerv1alpha1.AzureConfigList{}
		err := crdCLient.List(ctx, &crdList)
		if err == nil {
			anySuccess = true
			for _, azureConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if azureConfig.DeletionTimestamp == nil {
					clusterList = append(clusterList, Cluster{azureConfig.Name, azureConfig.Namespace, azure})
				}
			}
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing AzureConfigs: %s", err))
		}
	}

	// KVM
	{
		crdList := providerv1alpha1.KVMConfigList{}
		err := crdCLient.List(ctx, &crdList)
		if err == nil {
			anySuccess = true
			for _, kvmConfig := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				if kvmConfig.DeletionTimestamp == nil {
					clusterList = append(clusterList, Cluster{kvmConfig.Name, kvmConfig.Namespace, kvm})
				}
			}
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing KVMConfigs: %s", err))
		}
	}

	if !anySuccess {
		// No provider check was successful, raise an error.
		return clusterList, unableToGetTenantClustersError
	}

	// At least one provider check was successful (but possibly no tenant clusters were found).
	return clusterList, nil
}

func stringVersionCmp(versionStr string, def *semver.Version, reference *semver.Version) (bool, error) {
	var version *semver.Version
	var err error
	if versionStr == "" {
		version = def
	} else {
		version, err = semver.NewVersion(versionStr)
		if err != nil {
			return false, err
		}
	}

	if version.Compare(*reference) >= 0 {
		return true, nil
	}

	return false, nil
}

package giantnetes

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Utils struct {
	logger    micrologger.Logger
	K8sClient k8sclient.Interface
}

type clusterWithProvider struct {
	clusterID string
	provider  string
}

func NewUtils(logger micrologger.Logger, client k8sclient.Interface) (*Utils, error) {
	if logger == nil {
		return nil, microerror.Mask(microerror.New("logger can't be nil"))
	}
	if client == nil {
		return nil, microerror.Mask(microerror.New("client can't be nil"))
	}

	return &Utils{
		logger:    logger,
		K8sClient: client,
	}, nil
}

func (u *Utils) GetTenantClusters(ctx context.Context, backup v1alpha1.ETCDBackup) ([]ETCDInstance, error) {
	var instances []ETCDInstance
	crdClient := u.K8sClient.G8sClient()

	clusterList, err := u.getAllGuestClusters(ctx, u.K8sClient.G8sClient())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	u.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("Found %d tenant clusters", len(clusterList)))

	for _, cluster := range clusterList {
		u.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("Preparing instance entry for tenant clusters %s", cluster.clusterID))

		// check if the cluster release version has support for etcd backup
		versionSupported, err := u.checkClusterVersionSupport(cluster, crdClient)
		if err != nil {
			u.logger.Log("level", "error", "msg", "Failed to check release version for cluster "+cluster.clusterID, "reason", err)
			continue
		}
		if !versionSupported {
			u.logger.Log("level", "warning", "msg", "Cluster "+cluster.clusterID+" is too old for etcd backup. Skipping.")
			continue
		}

		// fetch etcd certs
		certs, err := u.fetchCerts(cluster.clusterID, u.K8sClient.K8sClient())
		if err != nil {
			u.logger.Log("level", "error", "msg", "Failed to fetch etcd certs for cluster "+cluster.clusterID, "reason", err)
			continue
		}
		// write etcd certs to tmpdir
		err = u.createCertFiles(cluster.clusterID, certs)
		if err != nil {
			u.logger.Log("level", "error", "msg", "Failed to write etcd certs to tmpdir for cluster "+cluster.clusterID, "reason", err)
			continue
		}

		// fetch etcd endpoint
		etcdEndpoint, err := u.getEtcdEndpoint(cluster, crdClient)
		if err != nil {
			u.logger.Log("level", "error", "msg", "Failed to fetch etcd endpoint for cluster "+cluster.clusterID, "reason", err)
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

// check if cluster release version has guest cluster backup support
func (u *Utils) checkClusterVersionSupport(cluster clusterWithProvider, crdCLient versioned.Interface) (bool, error) {
	getOpts := metav1.GetOptions{}

	switch cluster.provider {
	case aws:
		{
			crd, err := crdCLient.ProviderV1alpha1().AWSConfigs(crdNamespace).Get(cluster.clusterID, getOpts)
			if err != nil {
				return false, microerror.Maskf(err, "failed to get aws crd"+cluster.clusterID)
			}
			crdVersionStr := crd.Spec.VersionBundle.Version
			if crdVersionStr == "" {
				crdVersionStr = "0.0.0"
			}
			crdVersion := semver.New(crdVersionStr)
			if crdVersion.Compare(*awsSupportFrom) >= 0 {
				// version has support
				return true, nil
			} else {
				// version doesnt have support
				return false, nil
			}
		}
	case azure:
		{
			crd, err := crdCLient.ProviderV1alpha1().AzureConfigs(crdNamespace).Get(cluster.clusterID, getOpts)
			if err != nil {
				return false, microerror.Maskf(err, "failed to get azure crd "+cluster.clusterID)
			}
			crdVersionStr := crd.Spec.VersionBundle.Version
			if crdVersionStr == "" {
				crdVersionStr = "0.0.0"
			}
			crdVersion := semver.New(crdVersionStr)
			if crdVersion.Compare(*azureSupportFrom) >= 0 {
				// version has support
				return true, nil
			} else {
				// version doesnt have support
				return false, nil
			}
		}
	case kvm:
		{
			// kvm backups are always supported
			return true, nil
		}
	}
	return false, nil
}

// fetch etcd client certs
func (u *Utils) fetchCerts(clusterID string, k8sClient kubernetes.Interface) (*TLSClientConfig, error) {

	getOpts := metav1.GetOptions{}
	secret, err := k8sClient.CoreV1().Secrets(secretNamespace).Get(clusterID+"-etcd", getOpts)
	if err != nil {
		return nil, microerror.Maskf(err, "error getting etcd client certificates for guest cluster %s", clusterID)
	}

	certs := &TLSClientConfig{
		CAData:  secret.Data["ca"],
		KeyData: secret.Data["key"],
		CrtData: secret.Data["crt"],
	}

	return certs, nil
}

// fetch guest cluster etcd endpoint
func (u *Utils) getEtcdEndpoint(cluster clusterWithProvider, crdCLient versioned.Interface) (string, error) {
	getOpts := metav1.GetOptions{}
	var etcdEndpoint string

	switch cluster.provider {
	case aws:
		{
			crd, err := crdCLient.ProviderV1alpha1().AWSConfigs(crdNamespace).Get(cluster.clusterID, getOpts)
			if err != nil {
				return "", microerror.Maskf(err, "error getting aws crd for guest cluster %s", cluster.clusterID)
			}
			etcdEndpoint = AwsEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	case azure:
		{
			crd, err := crdCLient.ProviderV1alpha1().AzureConfigs(crdNamespace).Get(cluster.clusterID, getOpts)
			if err != nil {
				return "", microerror.Maskf(err, "error getting azure crd for guest cluster %s", cluster.clusterID)
			}
			etcdEndpoint = AzureEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	case kvm:
		{
			crd, err := crdCLient.ProviderV1alpha1().KVMConfigs(crdNamespace).Get(cluster.clusterID, getOpts)
			if err != nil {
				return "", microerror.Maskf(err, "error getting kvm crd for guest cluster %s", cluster.clusterID)
			}
			etcdEndpoint = KVMEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	}

	// we already check for unknown provider at the start
	return etcdEndpoint, nil
}

// create cert files in tmp dir from certConfig and saves filenames back
func (u *Utils) createCertFiles(clusterID string, certConfig *TLSClientConfig) error {
	tmpDir, err := ioutil.TempDir("", clusterID)
	if err != nil {
		return microerror.Mask(err)
	}

	// cert
	err = ioutil.WriteFile(CertFile(clusterID, tmpDir), certConfig.CrtData, fileMode)
	if err != nil {
		return microerror.Maskf(err, "Failed to write crt file "+CertFile(clusterID, tmpDir))
	}
	certConfig.CrtFile = CertFile(clusterID, tmpDir)

	// key
	err = ioutil.WriteFile(KeyFile(clusterID, tmpDir), certConfig.KeyData, fileMode)
	if err != nil {
		return microerror.Maskf(err, "Failed to write key file "+KeyFile(clusterID, tmpDir))
	}
	certConfig.KeyFile = KeyFile(clusterID, tmpDir)

	// ca
	err = ioutil.WriteFile(CAFile(clusterID, tmpDir), certConfig.CAData, fileMode)
	if err != nil {
		return microerror.Maskf(err, "Failed to write ca file "+CAFile(clusterID, tmpDir))
	}
	certConfig.CAFile = CAFile(clusterID, tmpDir)

	return nil
}

// fetch all guest clusters ids in host cluster
func (u *Utils) getAllGuestClusters(ctx context.Context, crdCLient versioned.Interface) ([]clusterWithProvider, error) {
	var clusterList []clusterWithProvider
	listOpt := metav1.ListOptions{}

	any := false

	// aws
	{
		crdList, err := crdCLient.ProviderV1alpha1().AWSConfigs(crdNamespace).List(listOpt)
		if err == nil {
			any = true
			for _, awsConfig := range crdList.Items {
				// only backup cluster if it was not marked for delete
				if awsConfig.DeletionTimestamp == nil {
					clusterList = append(clusterList, clusterWithProvider{awsConfig.Name, aws})
				}
			}
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing AWSConfigs: %s", err))
		}
	}

	// azure
	{
		crdList, err := crdCLient.ProviderV1alpha1().AzureConfigs(crdNamespace).List(listOpt)
		if err == nil {
			any = true
			for _, azureConfig := range crdList.Items {
				// only backup cluster if it was not marked for delete
				if azureConfig.DeletionTimestamp == nil {
					clusterList = append(clusterList, clusterWithProvider{azureConfig.Name, azure})
				}
			}
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing AzureConfigs: %s", err))
		}
	}

	// kvm
	{
		crdList, err := crdCLient.ProviderV1alpha1().KVMConfigs(crdNamespace).List(listOpt)
		if err == nil {
			any = true
			for _, kvmConfig := range crdList.Items {
				// only backup cluster if it was not marked for delete
				if kvmConfig.DeletionTimestamp == nil {
					clusterList = append(clusterList, clusterWithProvider{kvmConfig.Name, kvm})
				}
			}
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing KVMConfigs: %s", err))
		}
	}

	if !any {
		// no provider check was successful, raise an error
		return clusterList, unableToGetTenantClustersError
	}

	// at least one provider check was successful (but possibly no tenant clusters were found
	return clusterList, nil
}

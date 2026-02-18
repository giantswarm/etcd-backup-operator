package giantnetes

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/etcd-backup-operator/v4/pkg/etcd/proxy"
	"github.com/giantswarm/etcd-backup-operator/v4/service/controller/key"
)

const (
	certificateLabel      = "giantswarm.io/certificate"
	certificateLabelValue = "calico-etcd-client"

	skipEtcdBackupAnnotation = "giantswarm.io/etcd-backup-operator-skip-backup"
)

type Utils struct {
	logger    micrologger.Logger
	K8sClient k8sclient.Interface
}

type Cluster struct {
	clusterKey client.ObjectKey
	provider   string
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

func (u *Utils) GetTenantClusters(ctx context.Context) ([]ETCDInstance, error) {
	var instances []ETCDInstance

	clusterList, err := u.getAllWorkloadClusters(ctx, u.K8sClient.CtrlClient())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	u.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("Found %d tenant clusters", len(clusterList)))

	for _, cluster := range clusterList {
		u.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("Preparing instance entry for tenant clusters %s", cluster.clusterKey.Name))

		// Check if the cluster backup should be skipped
		backupSkipped, err := u.isClusterSkipped(ctx, cluster)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to check if backup should be skipped for cluster %s", cluster.clusterKey.Name), "reason", err)
			continue
		}
		if backupSkipped {
			u.logger.LogCtx(ctx, "level", "debug", "msg", fmt.Sprintf("Backup for cluster %s is skipped explicitly", cluster.clusterKey.Name))
			continue
		}

		// Check if the cluster release version has support for ETCD backup.
		versionSupported, err := u.checkClusterVersionSupport(ctx, cluster)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to check release version for cluster %s", cluster.clusterKey.Name), "reason", err)
			continue
		}
		if !versionSupported {
			u.logger.LogCtx(ctx, "level", "warning", "msg", fmt.Sprintf("Cluster %s is too old for etcd backup. Skipping.", cluster.clusterKey.Name))
			continue
		}

		// Prepare ETCD tls config.
		tlsConfig, err := u.getEtcdTLSCfg(ctx, cluster)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to fetch etcd certs for cluster %s", cluster.clusterKey.Name), "reason", err)
			continue
		}

		// Fetch ETCD endpoint.
		etcdEndpoint, err := u.getEtcdEndpoint(ctx, cluster)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to fetch etcd endpoint for cluster %s", cluster.clusterKey.Name), "reason", err)
			continue
		}

		// prepare etcd proxy
		p, err := u.getEtcdProxy(ctx, cluster, tlsConfig)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "error", "msg", fmt.Sprintf("Failed to prepare etcd proxy for cluster %s", cluster.clusterKey.Name), "reason", err)
			continue
		}

		instances = append(instances, ETCDInstance{
			Name:   cluster.clusterKey.Name,
			ETCDv2: ETCDv2Settings{},
			ETCDv3: ETCDv3Settings{
				Endpoints: etcdEndpoint,
				TLSConfig: tlsConfig,
				Proxy:     p,
			},
		})
	}
	return instances, nil
}

// isClusterSkipped checks if cluster should be skipped from guest cluster backup.
func (u *Utils) isClusterSkipped(ctx context.Context, cluster Cluster) (bool, error) {
	crdClient := u.K8sClient.CtrlClient()

	switch cluster.provider {
	case awsCAPI:
		crd := v1alpha3.AWSCluster{}
		err := crdClient.Get(ctx, cluster.clusterKey, &crd)
		if err != nil {
			return false, microerror.Maskf(executionFailedError, "error getting aws crd for guest cluster %#q with error %#q", cluster.clusterKey.Name, err)
		}

		if crd.Annotations[skipEtcdBackupAnnotation] == "true" {
			return true, nil
		}
	case azure:
		return false, nil
	case kvm:
		return false, nil
	case CAPI:
		return false, nil
	}
	return false, nil
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
			err := crdClient.Get(ctx, cluster.clusterKey, &crd)
			if err != nil {
				return false, microerror.Maskf(executionFailedError, "failed to get azure crd %#q with error %#q", cluster.clusterKey.Name, err)
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
				return false, microerror.Maskf(executionFailedError, "failed to get cluster version from AzureConfig %#q", cluster.clusterKey.Name)
			}
			return stringVersionCmp(version, semver.New("0.0.0"), azureSupportFrom)
		}
	case kvm:
		{
			// KVM backups are always supported.
			return true, nil
		}
	case CAPI:
		{
			// CAPI backups are always supported.
			return true, nil
		}
	}
	return false, nil
}

func (u *Utils) getEtcdTLSCfg(ctx context.Context, cluster Cluster) (*tls.Config, error) {
	if cluster.provider == CAPI {
		t, err := u.getCAPIEtcdTLSCfg(ctx, cluster.clusterKey)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		return t, nil
	} else {
		t, err := u.getLegacyEtcdTLSCfg(ctx, cluster.clusterKey)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		return t, nil
	}
}

// Fetch ETCD client certs.
func (u *Utils) getLegacyEtcdTLSCfg(ctx context.Context, clusterKey client.ObjectKey) (*tls.Config, error) {
	k8sClient := u.K8sClient.CtrlClient()
	secrets := v1.SecretList{}
	err := k8sClient.List(ctx, &secrets, client.MatchingLabels{
		label.Cluster:    clusterKey.Name,
		certificateLabel: certificateLabelValue,
	})
	if err != nil {
		return nil, microerror.Maskf(executionFailedError, "error getting etcd client certificates for guest cluster %#q with error %#q", clusterKey.Name, err)
	}

	if len(secrets.Items) != 1 {
		return nil, microerror.Maskf(executionFailedError, "expected exactly 1 secret with %s=%q and %s=%q, got %d", label.Cluster, clusterKey.Name, certificateLabel, certificateLabelValue, len(secrets.Items))
	}

	s := secrets.Items[0]

	tlsConfig, err := key.PrepareTLSConfig(s.Data["ca"], s.Data["crt"], s.Data["key"])
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tlsConfig, nil
}

// Fetch ETCD client certs.
func (u *Utils) getEtcdProxy(ctx context.Context, cluster Cluster, tlsConfig *tls.Config) (*proxy.Proxy, error) {
	if cluster.provider == CAPI {
		ctrClient := u.K8sClient.CtrlClient()

		targetClusterRESTConfig, err := key.RESTConfig(ctx, ctrClient, cluster.clusterKey)
		if err != nil {
			return nil, microerror.Maskf(executionFailedError, "error fetching CAPI cluster rest config for cluster  %#q with error %#q", cluster.clusterKey.Name, err)
		}

		p := &proxy.Proxy{
			Kind:       "pods",
			Namespace:  metav1.NamespaceSystem,
			KubeConfig: targetClusterRESTConfig,
			TLSConfig:  tlsConfig,
			Port:       2379,
		}

		return p, nil

	} else {
		// no proxy needed for legacy clusters
		return nil, nil
	}
}

// Fetch ETCD client certs for CAPI cluster.
func (u *Utils) getCAPIEtcdTLSCfg(ctx context.Context, clusterKey client.ObjectKey) (*tls.Config, error) {
	ctrlClient := u.K8sClient.CtrlClient()

	etcdCerts := &v1.Secret{}
	etcdCertsObjectKey := client.ObjectKey{
		Namespace: clusterKey.Namespace,
		Name:      fmt.Sprintf("%s-etcd", clusterKey.Name),
	}
	if err := ctrlClient.Get(ctx, etcdCertsObjectKey, etcdCerts); err != nil {
		return nil, microerror.Mask(err)
	}
	crtData, ok := etcdCerts.Data[secret.TLSCrtDataName]
	if !ok {
		return nil, microerror.Maskf(executionFailedError, "etcd tls crt does not exist for cluster %s/%s", clusterKey.Namespace, clusterKey.Name)
	}
	keyData, ok := etcdCerts.Data[secret.TLSKeyDataName]
	if !ok {
		return nil, microerror.Maskf(executionFailedError, "etcd tls key does not exist for cluster %s/%s", clusterKey.Namespace, clusterKey.Name)
	}

	tlsConfig, err := key.PrepareTLSConfig(crtData, crtData, keyData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tlsConfig, nil
}

// Fetch guest cluster ETCD endpoint.
func (u *Utils) getEtcdEndpoint(ctx context.Context, cluster Cluster) (string, error) {
	var etcdEndpoint string
	crdClient := u.K8sClient.CtrlClient()

	switch cluster.provider {
	case awsCAPI:
		{
			crd := v1alpha3.AWSCluster{}
			err := crdClient.Get(ctx, cluster.clusterKey, &crd)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error getting aws crd for guest cluster %#q with error %#q", cluster.clusterKey.Name, err)
			}
			if crd.Spec.Cluster.DNS.Domain == "" {
				return "", microerror.Maskf(executionFailedError, "awscluster %#q does not have any cluster domain set in spec.cluster.dns.domain", cluster.clusterKey.Name)
			}
			etcdEndpoint = AwsCAPIEtcdEndpoint(cluster.clusterKey.Name, crd.Spec.Cluster.DNS.Domain)
			break
		}
	case azure:
		{
			crd := providerv1alpha1.AzureConfig{}
			err := crdClient.Get(ctx, cluster.clusterKey, &crd)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error getting azure crd for guest cluster %#q with error %#q", cluster.clusterKey.Name, err)
			}
			etcdEndpoint = AzureEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	case kvm:
		{
			crd := providerv1alpha1.KVMConfig{}
			err := crdClient.Get(ctx, cluster.clusterKey, &crd)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error getting kvm crd for guest cluster %#q with error %#q", cluster.clusterKey.Name, err)
			}
			etcdEndpoint = KVMEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	case CAPI:
		{
			// for CAPI endpoint we need to fetch workload cluster k8s client and look for etcd pods
			targetClusterRESTConfig, err := key.RESTConfig(ctx, crdClient, cluster.clusterKey)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error fetching CAPI cluster rest config for cluster  %#q with error %#q", cluster.clusterKey.Name, err)
			}

			targetCtrlClient, err := key.GetCtrlClient(targetClusterRESTConfig)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error creating CAPI cluster kubernetes client for cluster  %#q with error %#q", cluster.clusterKey.Name, err)
			}

			// Specify the label for etcd pods
			labelSelector := client.MatchingLabels(map[string]string{EtcdLabelComponentKey: EtcdLabelComponentValue, EtcdLabelTierKey: EtcdLabelTierValue})

			// List pods with the specified label
			podList := v1.PodList{}
			err = targetCtrlClient.List(ctx, &podList, labelSelector)
			if err != nil {
				return "", microerror.Maskf(executionFailedError, "error creating CAPI cluster kubernetes client for cluster  %#q with error %#q", cluster.clusterKey.Name, err)
			}

			if len(podList.Items) == 0 {
				return "", microerror.Maskf(executionFailedError, "error getting etcd endpoint, no etcd pods found in cluster  %#q with error %#q", cluster.clusterKey.Name, err)
			}

			etcdEndpoint = podList.Items[0].Name
			break
		}
	}

	// We already check for unknown provider at the start.
	return etcdEndpoint, nil
}

// Fetch all workload clusters IDs in host cluster.
func (u *Utils) getAllWorkloadClusters(ctx context.Context, crdCLient client.Client) ([]Cluster, error) {
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
					clusterList = append(clusterList, Cluster{clusterKey: client.ObjectKey{Name: awsClusterObj.Name, Namespace: awsClusterObj.Namespace}, provider: awsCAPI})
				}
			}
		} else if isMissingCRDError(err) {
			// ignore missing CRD/KIND error as its expected that single MC do not have all provider CRs
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
					clusterList = append(clusterList, Cluster{clusterKey: client.ObjectKey{Name: azureConfig.Name, Namespace: azureConfig.Namespace}, provider: azure})
				}
			}
		} else if isMissingCRDError(err) {
			// ignore missing CRD/KIND error as its expected that single MC do not have all provider CRs
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
					clusterList = append(clusterList, Cluster{clusterKey: client.ObjectKey{Name: kvmConfig.Name, Namespace: kvmConfig.Namespace}, provider: kvm})
				}
			}
		} else if isMissingCRDError(err) {
			// ignore missing CRD/KIND error as its expected that single MC do not have all provider CRs
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing KVMConfigs: %s", err))
		}
	}

	// CAPI
	{
		crdList := capi.ClusterList{}
		err := crdCLient.List(ctx, &crdList)
		if err == nil {
			anySuccess = true
			for _, cluster := range crdList.Items {
				// Only backup cluster if it was not marked for delete.
				// and if the control and infrastructure is ready
				if cluster.DeletionTimestamp == nil &&
				cluster.Status.Initialization.ControlPlaneInitialized != nil && *cluster.Status.Initialization.ControlPlaneInitialized &&
				cluster.Status.Initialization.InfrastructureProvisioned != nil && *cluster.Status.Initialization.InfrastructureProvisioned {
					clusterList = append(clusterList, Cluster{clusterKey: client.ObjectKey{Name: cluster.Name, Namespace: cluster.Namespace}, provider: CAPI})
				}
			}
		} else {
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Error listing CAPI Clusters: %s", err))
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

func isMissingCRDError(err error) bool {
	return strings.Contains(err.Error(), "no matches for kind")
}

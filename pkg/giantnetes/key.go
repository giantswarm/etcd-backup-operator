package giantnetes

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
)

const (
	// CRD names for each provider.
	awsCAPI = "awsCAPI"
	azure   = "azure"
	kvm     = "kvm"
	CAPI    = "capi"
)

const (
	EtcdLabelComponentKey   = "component"
	EtcdLabelComponentValue = "etcd"
	EtcdLabelTierKey        = "tier"
	EtcdLabelTierValue      = "control-plane"
)

var azureSupportFrom *semver.Version = semver.Must(semver.NewVersion("0.2.0"))

func AwsCAPIEtcdEndpoint(clusterID string, baseDomain string) string {
	return fmt.Sprintf("https://etcd.%s.k8s.%s:2379", clusterID, baseDomain)
}

func AzureEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:2379", etcdDomain)
}

func KVMEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:443", etcdDomain)
}

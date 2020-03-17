package giantnetes

import (
	"fmt"
	"path"

	"github.com/coreos/go-semver/semver"
)

const (
	// CRD names for each provider.
	aws     = "aws"
	awsCAPI = "awsCAPI"
	azure   = "azure"
	kvm     = "kvm"

	// Namespace where secrets are located.
	secretNamespace = "default"

	// Namespace where CRDs are located.
	crdNamespace = "default"

	fileMode = 0600
)

var awsSupportFrom *semver.Version = semver.Must(semver.NewVersion("3.1.1"))
var azureSupportFrom *semver.Version = semver.Must(semver.NewVersion("0.2.0"))

func AwsEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:2379", etcdDomain)
}

func AwsCAPIEtcdEndpoint(clusterID string, baseDomain string) string {
	return fmt.Sprintf("https://etcd.%s.k8s.%s:2379", clusterID, baseDomain)
}

func AzureEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:2379", etcdDomain)
}

func BackupPrefix(clusterID string) string {
	return "-" + clusterID
}

func CAFile(clusterID string, tmpDir string) string {
	return path.Join(tmpDir, fmt.Sprintf("%s-%s.pem", clusterID, "ca"))
}

func CertFile(clusterID string, tmpDir string) string {
	return path.Join(tmpDir, fmt.Sprintf("%s-%s.pem", clusterID, "crt"))
}

func KeyFile(clusterID string, tmpDir string) string {
	return path.Join(tmpDir, fmt.Sprintf("%s-%s.pem", clusterID, "key"))
}

func KVMEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:443", etcdDomain)
}

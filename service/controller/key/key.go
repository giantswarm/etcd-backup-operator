package key

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	backupv1alpha1 "github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kcfg "sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ManagementCluster = "ManagementCluster"

	// Environment variables.
	EnvAWSAccessKeyID     = "AWS_ACCESS_KEY_ID"
	EnvAWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY" // nolint: gosec
	EncryptionPassword    = "ENCRYPTION_PASSWORD"
)

func ToCustomObject(v interface{}) (backupv1alpha1.ETCDBackup, error) {
	if v == nil {
		return backupv1alpha1.ETCDBackup{}, microerror.Maskf(executionFailedError, "expected '%T', got '%T'", &backupv1alpha1.ETCDBackup{}, v)
	}

	customObjectPointer, ok := v.(*backupv1alpha1.ETCDBackup)
	if !ok {
		return backupv1alpha1.ETCDBackup{}, microerror.Maskf(executionFailedError, "expected '%T', got '%T'", &backupv1alpha1.ETCDBackup{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func FilenamePrefix(installationName string, clusterName string) string {
	return fmt.Sprintf("%s-%s", installationName, clusterName)
}

func PrepareTLSConfig(caData []byte, crtData []byte, keyData []byte) (*tls.Config, error) {
	clientCert, err := tls.X509KeyPair(crtData, keyData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caData)
	tlsConfig := &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
	}
	tlsConfig.InsecureSkipVerify = true

	return tlsConfig, nil
}

func TLSConfigFromCertFiles(ca string, cert string, key string) (*tls.Config, error) {
	caData, err := os.ReadFile(ca) //nolint:gosec
	if err != nil {
		return nil, microerror.Mask(err)
	}

	crtData, err := os.ReadFile(cert) //nolint:gosec
	if err != nil {
		return nil, microerror.Mask(err)
	}

	keyData, err := os.ReadFile(key) //nolint:gosec
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tlsCofig, err := PrepareTLSConfig(caData, crtData, keyData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tlsCofig, nil
}

// RESTConfig returns a configuration instance to be used with a Kubernetes client.
func RESTConfig(ctx context.Context, c client.Reader, cluster client.ObjectKey) (*restclient.Config, error) {
	kubeConfig, err := kcfg.FromSecret(ctx, c, cluster)
	if err != nil {
		return nil, microerror.Maskf(executionFailedError, "failed to retrieve kubeconfig secret for Cluster %s/%s : %s", cluster.Namespace, cluster.Name, err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return nil, microerror.Maskf(executionFailedError, "failed to create REST configuration for Cluster %s/%s : %s", cluster.Namespace, cluster.Name, err)
	}

	return restConfig, nil
}

func GetCtrlClient(config *restclient.Config) (client.Client, error) {
	s := runtime.NewScheme()

	schemes := []func(*runtime.Scheme) error{
		corev1.AddToScheme,
	}

	// Extend the global client-go scheme which is used by all the tools under
	// the hood. The scheme is required for the controller-runtime controller to
	// be able to watch for runtime objects of a certain type.
	schemeBuilder := runtime.SchemeBuilder(schemes)

	err := schemeBuilder.AddToScheme(s)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	c, err := client.New(config, client.Options{Scheme: s})
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return c, nil
}

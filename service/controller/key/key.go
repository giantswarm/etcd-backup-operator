package key

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	backupv1alpha1 "github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	"github.com/giantswarm/microerror"
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
	caData, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	crtData, err := ioutil.ReadFile(cert)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	keyData, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tlsCofig, err := PrepareTLSConfig(caData, crtData, keyData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tlsCofig, nil
}

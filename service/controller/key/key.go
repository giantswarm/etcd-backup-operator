package key

import (
	"fmt"

	backupv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/backup/v1alpha1"
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

package key

import (
	"fmt"
	"os"

	backupv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	ControlPlane = "Control Plane"

	EnvFilenamePrefix = "FILENAME_PREFIX"
	DefaultPrefix     = "etcd-backup"

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

func FilenamePrefix(instanceName string) string {
	globalPrefix := os.Getenv(EnvFilenamePrefix)
	if len(globalPrefix) == 0 {
		globalPrefix = DefaultPrefix
	}

	if instanceName == ControlPlane {
		return globalPrefix
	}
	return fmt.Sprintf("%s-%s", globalPrefix, instanceName)
}

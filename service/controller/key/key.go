package key

import (
	"fmt"
	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	"os"
)

const (
	StatusEmpty     = ""
	StatusPending   = "Pending"
	StatusRunning   = "Running"
	StatusFailed    = "Failed"
	StatusCompleted = "Completed"

	AllowedBackupAttempts = int8(3)
	CRKeepTimeoutSeconds  = 7 * 24 * 60 * 60

	ControlPlane = "Control Plane"

	EnvFilenamePrefix = "FILENAME_PREFIX"
	DefaultPrefix     = "etcd-backup"
)

func ToETCDBackup(v interface{}) (backupv1alpha1.ETCDBackup, error) {
	if v == nil {
		return backupv1alpha1.ETCDBackup{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &backupv1alpha1.ETCDBackup{}, v)
	}

	customObjectPointer, ok := v.(*backupv1alpha1.ETCDBackup)
	if !ok {
		return backupv1alpha1.ETCDBackup{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &backupv1alpha1.ETCDBackup{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func GetPrefix(instanceName string) string {
	globalPrefix := os.Getenv(EnvFilenamePrefix)
	if len(globalPrefix) == 0 {
		globalPrefix = DefaultPrefix
	}

	if instanceName == ControlPlane {
		return globalPrefix
	}
	return fmt.Sprintf("%s-%s", globalPrefix, instanceName)
}

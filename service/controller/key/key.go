package key

import (
	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
)

func ToCustomObject(v interface{}) (backupv1alpha1.ETCDBackup, error) {
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

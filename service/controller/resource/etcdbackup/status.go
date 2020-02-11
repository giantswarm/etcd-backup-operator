package etcdbackup

import (
	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
)

func (r *Resource) getGlobalStatus(customObject backupv1alpha1.ETCDBackup) (string, error) {
	return customObject.Status.Status, nil
}

func (r *Resource) setGlobalStatus(customObject backupv1alpha1.ETCDBackup, updatedStatus string) error {
	customObject.Status.Status = updatedStatus

	return r.persistCustomObject(customObject)
}

func (r *Resource) persistCustomObject(customObject backupv1alpha1.ETCDBackup) error {
	_, err := r.k8sClient.G8sClient().BackupV1alpha1().ETCDBackups().UpdateStatus(&customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

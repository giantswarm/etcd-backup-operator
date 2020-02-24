package etcdbackup

import (
	"context"
	"fmt"
	"time"

	backupv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/pkg/giantnetes"
)

func (r *Resource) getGlobalStatus(customObject backupv1alpha1.ETCDBackup) (string, error) {
	return customObject.Status.Status, nil
}

func (r *Resource) setGlobalStatus(customObject backupv1alpha1.ETCDBackup, updatedStatus string) error {
	customObject.Status.Status = updatedStatus

	return r.persistCustomObject(customObject)
}

func (r *Resource) findOrInitializeInstanceStatus(ctx context.Context, etcdBackup backupv1alpha1.ETCDBackup, instance giantnetes.ETCDInstance) backupv1alpha1.ETCDInstanceBackupStatusIndex {
	status, found := etcdBackup.Status.Instances[instance.Name]
	if found {
		return status
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Initializing new ETCDInstanceBackupStatus for %s", instance.Name))

	now := time.Now().UTC()

	newStatus := backupv1alpha1.ETCDInstanceBackupStatusIndex{
		Name: instance.Name,
		V2: backupv1alpha1.ETCDInstanceBackupStatus{
			Status:   InstanceBackupStatePending,
			Attempts: 0,
			StartedTimestamp: backupv1alpha1.DeepCopyTime{
				Time: now,
			},
		},
		V3: backupv1alpha1.ETCDInstanceBackupStatus{
			Status:   InstanceBackupStatePending,
			Attempts: 0,
			StartedTimestamp: backupv1alpha1.DeepCopyTime{
				Time: now,
			},
		},
	}

	return newStatus
}

func isTerminalInstaceState(state string) bool {
	return state == InstanceBackupStateCompleted || state == InstanceBackupStateFailed
}

func (r *Resource) setInstanceV2Status(ctx context.Context, customObject backupv1alpha1.ETCDBackup, instanceName string, newStatus string) error {
	status, found := customObject.Status.Instances[instanceName]
	if !found {
		return microerror.Mask(microerror.Newf("Instances status was unexpectedly missing for %s", instanceName))
	}

	status.V2.Status = newStatus
	customObject.Status.Instances[instanceName] = status

	return r.persistCustomObject(customObject)
}

func (r *Resource) setInstanceV3Status(ctx context.Context, customObject backupv1alpha1.ETCDBackup, instanceName string, newStatus string) error {
	status, found := customObject.Status.Instances[instanceName]
	if !found {
		return microerror.Mask(microerror.Newf("Instances status was unexpectedly missing for %s", instanceName))
	}

	status.V3.Status = newStatus
	customObject.Status.Instances[instanceName] = status

	return r.persistCustomObject(customObject)
}

func (r *Resource) persistCustomObject(customObject backupv1alpha1.ETCDBackup) error {
	_, err := r.k8sClient.G8sClient().BackupV1alpha1().ETCDBackups().UpdateStatus(&customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

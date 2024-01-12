package etcdbackup

import (
	"context"
	"fmt"

	backupv1alpha1 "github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/v4/service/controller/resource/etcdbackup/internal/state"
)

// Sets the initial state.
func (r *Resource) backupEmptyTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	var err error
	r.logger.LogCtx(ctx, "level", "debug", "message", "no current state present")

	cr, ok := obj.(*backupv1alpha1.ETCDBackup)
	if !ok {
		return instanceBackupStateFailed, microerror.Mask(fmt.Errorf("expected v1alpha1.EtcdBackup, got %T", obj))
	}

	backups := backupv1alpha1.ETCDBackupList{}
	err = r.k8sClient.CtrlClient().List(ctx, &backups)
	if err != nil {
		return instanceBackupStateSkipped, microerror.Mask(err)
	}

	var latestBackup backupv1alpha1.ETCDBackup
	for _, backup := range backups.Items {
		if latestBackup.Name < backup.Name {
			latestBackup = backup
		}
	}

	if cr.Name != latestBackup.Name {
		r.logger.LogCtx(ctx, "level", "debug", "message", "Backup is not the latest, skipping")
		return instanceBackupStateSkipped, nil
	}

	if latestBackup.Status.FinishedTimestamp.IsZero() {
		r.logger.LogCtx(ctx, "level", "debug", "message", "Backup is already running, skipping")
		return instanceBackupStateSkipped, nil
	}

	return backupStatePending, nil
}

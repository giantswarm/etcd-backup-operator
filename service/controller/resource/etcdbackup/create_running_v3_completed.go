package etcdbackup

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

func (r *Resource) backupRunningV3BackupCompletedTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	// Check if any of the instances failed, and in that case set the Backup state to Failed.
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	for _, i := range customObject.Status.Instances {
		if i.V2.Status == instanceBackupStateFailed || i.V3.Status == instanceBackupStateFailed {
			return backupStateFailed, nil
		}
	}

	return backupStateCompleted, nil
}
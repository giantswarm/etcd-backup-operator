package etcdbackup

import (
	"context"

	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

func (r *Resource) backupRunningV3BackupCompletedTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	// TODO Check if any of the instances failed, and in that case set the Backup state to Failed.
	return backupStateCompleted, nil
}

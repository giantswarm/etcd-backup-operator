package etcdbackup

import (
	"context"

	"github.com/giantswarm/etcd-backup-operator/v4/service/controller/resource/etcdbackup/internal/state"
)

func (r *Resource) backupRunningV2BackupCompletedTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	return backupStateRunningV3BackupRunning, nil
}

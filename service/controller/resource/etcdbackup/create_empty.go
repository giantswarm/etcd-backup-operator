package etcdbackup

import (
	"context"

	"github.com/giantswarm/etcd-backup-operator/v3/service/controller/resource/etcdbackup/internal/state"
)

// Sets the initial state.
func (r *Resource) backupEmptyTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", "no current state present")

	return backupStatePending, nil
}

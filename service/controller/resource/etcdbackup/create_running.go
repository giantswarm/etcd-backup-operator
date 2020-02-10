package etcdbackup

import (
	"context"

	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

func (r *Resource) runningTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	return "", nil
}

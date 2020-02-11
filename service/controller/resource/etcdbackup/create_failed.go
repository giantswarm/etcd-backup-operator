package etcdbackup

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

// Deletes the ETCDBackup if it's older than the threshold.
func (r *Resource) globalBackupFailedTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	err = r.cleanup(ctx, customObject)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return GlobalBackupSateFailed, nil
}

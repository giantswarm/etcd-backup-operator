package etcdbackup

import (
	"context"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

// Sets the StartedTimestamp for the global reconciliation and initializes the Status->Instances field.
// Then, it moves to the Running stage.
func (r *Resource) backupPendingTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Initializing global Status")
	customObject.Status.StartedTimestamp = v1alpha1.DeepCopyTime{
		Time: time.Now().UTC(),
	}

	err = r.persistCustomObjectStatus(customObject)
	if err != nil {
		return "", microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Initialized global Status")

	// No need to cancel the reconciliation: the state is changing so this will be done in EnsureCreated.

	return backupStateRunningV2BackupRunning, nil
}

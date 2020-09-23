package etcdbackup

import (
	"context"
	"time"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/etcd-backup-operator/v2/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/v2/service/controller/resource/etcdbackup/internal/state"
)

// Sets the StartedTimestamp for the global reconciliation and initializes the Status->Instances field.
// Then, it moves to the Running stage.
func (r *Resource) backupPendingTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Initializing global Status")
	customObject.Status.StartedTimestamp = metav1.Time{Time: time.Now().UTC()}

	err = r.persistCustomObjectStatus(ctx, customObject)
	if err != nil {
		return "", microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Initialized global Status")

	// No need to cancel the reconciliation: the state is changing so this will be done in EnsureCreated.

	return backupStateRunningV2BackupRunning, nil
}

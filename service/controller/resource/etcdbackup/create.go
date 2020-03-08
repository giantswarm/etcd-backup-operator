package etcdbackup

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

const (
	// Global States.
	backupStateEmpty                    = ""
	backupStatePending                  = "Pending"
	backupStateRunningV2BackupRunning   = "V2BackupRunning"
	backupStateRunningV2BackupCompleted = "V2BackupCompleted"
	backupStateRunningV3BackupRunning   = "V3BackupRunning"
	backupStateRunningV3BackupCompleted = "V3BackupCompleted"
	backupStateCompleted                = "Completed"
	backupStateFailed                   = "Failed"

	// Instance States.
	instanceBackupStatePending   = "Pending"
	instanceBackupStateCompleted = "Completed"
	instanceBackupStateFailed    = "Failed"
	instanceBackupStateSkipped   = "Skipped"

	// Various settings.
	maxBackupAttempts = int8(3)

	// Default values.
	crKeepTimeoutSeconds = 7 * 24 * 60 * 60
)

// configureStateMachine configures and returns state machine that is driven by
// EnsureCreated.
func (r *Resource) configureStateMachine() {
	sm := state.Machine{
		backupStateEmpty:                    r.backupEmptyTransition,
		backupStatePending:                  r.backupPendingTransition,
		backupStateRunningV2BackupRunning:   r.backupRunningV2BackupRunningTransition,
		backupStateRunningV2BackupCompleted: r.backupRunningV2BackupCompletedTransition,
		backupStateRunningV3BackupRunning:   r.backupRunningV3BackupRunningTransition,
		backupStateRunningV3BackupCompleted: r.backupRunningV3BackupCompletedTransition,
		backupStateCompleted:                r.backupCompletedTransition,
		backupStateFailed:                   r.backupFailedTransition,
	}

	r.stateMachine = sm
}

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var newState state.State
	var currentState state.State
	{
		s, err := r.getGlobalStatus(customObject)
		if err != nil {
			return microerror.Mask(err)
		}
		currentState = state.State(s)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("current state: %s", currentState))
		newState, err = r.stateMachine.Execute(ctx, obj, currentState)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if newState != currentState {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("new state: %s", newState))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting resource status to '%s'", newState))
		err = r.setGlobalStatus(customObject, string(newState))
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status to '%s'", newState))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no state change")
	}

	return nil
}

package etcdbackup

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/etcd-backup-operator/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

func (r *Resource) doNothingTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	// Todo this transition does nothing and will be implemented in a future PR
	return "", nil
}

// This function handles the reconciliation state machine for a specific ETCD version (v2 or v3) of a single ETCD instance within the global reconciliation loop.
func (r *Resource) runInstanceStateMachine(ctx context.Context, obj v1alpha1.ETCDBackup, status v1alpha1.ETCDInstanceBackupStatus) (state.State, error) {
	stateMachine := state.Machine{
		InstanceBackupStateEmpty:     r.doNothingTransition,
		InstanceBackupStatePending:   r.doNothingTransition,
		InstanceBackupStateRunning:   r.doNothingTransition,
		InstanceBackupStateCompleted: r.doNothingTransition,
		InstanceBackupStateFailed:    r.doNothingTransition,
	}
	var err error
	var newInstanceState state.State
	currentInstanceState := state.State(status.Status)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("current state: %s", currentInstanceState))
	newInstanceState, err = stateMachine.Execute(ctx, obj, currentInstanceState)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return newInstanceState, nil
}

func (r *Resource) globalBackupRunningTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	anyWipInstance := false
	anyFailedInstance := false

	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	utils, err := giantnetes.NewUtils(r.logger, r.k8sClient)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// control plane
	instances := []giantnetes.ETCDInstance{
		{
			Name:   "ControlPlane",
			ETCDv2: r.etcdV2Settings,
			ETCDv3: r.etcdV3Settings,
		},
	}

	if customObject.Spec.GuestBackup {
		// tenant clusters
		guestInstances, err := utils.GetTenantClusters(ctx, customObject)
		if err != nil {
			return "", microerror.Mask(err)
		}
		instances = append(instances, guestInstances...)
	}

	for _, etcdinstance := range instances {
		instanceStatus := r.findOrInitializeInstanceStatus(ctx, customObject, etcdinstance)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting working on instance %s", instanceStatus.Name))

		// Run the state machine step for V2 ETCD.
		{
			newV2Status, err := r.runInstanceStateMachine(ctx, customObject, instanceStatus.V2)
			if err != nil {
				return "", microerror.Mask(err)
			}

			if newV2Status == InstanceBackupStateFailed {
				anyFailedInstance = true
			}

			if newV2Status != state.State(instanceStatus.V2.Status) {
				anyWipInstance = true
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("new state: %s", newV2Status))
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting instance status to '%s'", newV2Status))
				err = r.setInstanceV3Status(ctx, customObject, etcdinstance.Name, string(newV2Status))
				if err != nil {
					return "", microerror.Mask(err)
				}
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status to '%s'", etcdinstance.Name))
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
				reconciliationcanceledcontext.SetCanceled(ctx)
				continue
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", "no state change")
			}
		}

		// Run the state machine step for V3 ETCD.
		{
			newV3Status, err := r.runInstanceStateMachine(ctx, customObject, instanceStatus.V3)
			if err != nil {
				return "", microerror.Mask(err)
			}

			if newV3Status == InstanceBackupStateFailed {
				anyFailedInstance = true
			}

			if newV3Status != state.State(instanceStatus.V3.Status) {
				anyWipInstance = true
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("new state: %s", newV3Status))
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting instance status to '%s'", newV3Status))
				err = r.setInstanceV3Status(ctx, customObject, etcdinstance.Name, string(newV3Status))
				if err != nil {
					return "", microerror.Mask(err)
				}
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status to '%s'", etcdinstance.Name))
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
				reconciliationcanceledcontext.SetCanceled(ctx)
				continue
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", "no state change")
			}
		}
	}

	if anyWipInstance {
		return GlobalBackupStateRunning, nil
	} else {
		if anyFailedInstance {
			return GlobalBackupStateFailed, nil
		} else {
			return GlobalBackupStateCompleted, nil
		}
	}
}

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

func (r *Resource) backupRunningV2BackupRunningTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	utils, err := giantnetes.NewUtils(r.logger, r.k8sClient)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Control plane.
	instances := []giantnetes.ETCDInstance{
		{
			Name:   "ControlPlane",
			ETCDv2: r.etcdV2Settings,
			ETCDv3: r.etcdV3Settings,
		},
	}

	if customObject.Spec.GuestBackup {
		// Tenant clusters.
		guestInstances, err := utils.GetTenantClusters(ctx, customObject)
		if err != nil {
			return "", microerror.Mask(err)
		}
		instances = append(instances, guestInstances...)
	}

	for _, etcdinstance := range instances {
		instanceStatus := r.findOrInitializeInstanceStatus(ctx, customObject, etcdinstance)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting working on instance %s", etcdinstance))

		newStatus, err := r.performETCDv2Backup(ctx, etcdinstance.ETCDv2, instanceStatus.V2)
		if err != nil {
			return "", microerror.Mask(err)
		}

		if newStatus != instanceStatus.V2.Status {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("new state: %s", newStatus))
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting instance status to '%s'", newStatus))
			err = r.setInstanceV2Status(ctx, customObject, etcdinstance.Name, string(newStatus))
			if err != nil {
				return "", microerror.Mask(err)
			}
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status to '%s'", etcdinstance.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return BackupStateRunningV2BackupRunning, nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "no state change")
		}
	}

	// No status changes have happened within any of the instances, backup is completed.
	return BackupStateRunningV2BackupCompleted, nil
}

func (r *Resource) performETCDv2Backup(ctx context.Context, etcdinstance giantnetes.ETCDv2Settings, status v1alpha1.ETCDInstanceBackupStatus) (string, error) {
	// If state is terminal, there's nothing else we can do on this instance, so just return the current state.
	if isTerminalInstaceState(status.Status) {
		return status.Status, nil
	}

	// TODO Try to do the backup.

	return InstanceBackupStateCompleted, nil
}

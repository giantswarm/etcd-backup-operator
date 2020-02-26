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

func (r *Resource) backupRunningV3BackupRunningTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
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
			Name:   key.ControlPlane,
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

	for _, etcdInstance := range instances {
		instanceStatus := r.findOrInitializeInstanceStatus(ctx, customObject, etcdInstance)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting working on instance %s", etcdInstance))

		newStatus, err := r.performETCDv3Backup(ctx, etcdInstance.ETCDv3, instanceStatus.V3)

		if newStatus != instanceStatus.V3.Status {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("new state: %s", newStatus))
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting instance status to '%s'", newStatus))
			err = r.setInstanceV3Status(ctx, customObject, etcdInstance.Name, string(newStatus))
			if err != nil {
				return "", microerror.Mask(err)
			}
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status to '%s'", etcdInstance.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return backupStateRunningV3BackupRunning, nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "no state change")
		}
	}

	// No status changes have happened within any of the instances, backup is completed.
	return backupStateRunningV3BackupCompleted, nil
}

func (r *Resource) performETCDv3Backup(ctx context.Context, etcdinstance giantnetes.ETCDv3Settings, status v1alpha1.ETCDInstanceBackupStatus) (string, error) {
	// If state is terminal, there's nothing else we can do on this instance, so just return the current state.
	if isTerminalInstaceState(status.Status) {
		return status.Status, nil
	}

	// TODO Try to do the backup.

	return instanceBackupStateCompleted, nil
}

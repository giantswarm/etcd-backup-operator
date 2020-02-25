package etcdbackup

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd"
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

		if etcdInstance.ETCDv3.AreComplete() {
			// If state is terminal, there's nothing else we can do on this instance, so just skip to next one.
			if isTerminalInstaceState(instanceStatus.V3.Status) {
				continue
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting v3 backup on instance %s", etcdInstance.Name))

			backupper := etcd.V3Backupper{
				CACert:    etcdInstance.ETCDv3.CaCert,
				Cert:      etcdInstance.ETCDv3.Cert,
				EncPass:   os.Getenv("ENCRYPTION_PASSWORD"),
				Endpoints: etcdInstance.ETCDv3.Endpoints,
				Logger:    r.logger,
				Key:       etcdInstance.ETCDv3.Key,
				Prefix:    key.GetPrefix(instanceStatus.Name),
			}

			err := r.performBackup(ctx, backupper, instanceStatus.Name)
			if err == nil {
				// Backup was successful.
				instanceStatus.V3.LatestError = ""
				instanceStatus.V3.Status = InstanceBackupStateCompleted
			} else {
				// Backup was unsuccessful.
				instanceStatus.V3.LatestError = err.Error()
				instanceStatus.V3.Status = InstanceBackupStateFailed
			}

			instanceStatus.V3.FinishedTimestamp = v1alpha1.DeepCopyTime{
				Time: time.Now().UTC(),
			}
		} else {
			r.logger.LogCtx(ctx, "level", "info", "message", "V3 backup skipped for %s because ETCD V3 setting are not set.", etcdInstance.Name)
			instanceStatus.V3.Status = InstanceBackupStateSkipped
		}

		customObject.Status.Instances[etcdInstance.Name] = instanceStatus

		err = r.persistCustomObject(customObject)
		if err != nil {
			return "", microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status to '%s'", etcdInstance.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return BackupStateRunningV3BackupRunning, nil
	}

	// No status changes have happened within any of the instances, backup is completed.
	return backupStateRunningV3BackupCompleted, nil
}

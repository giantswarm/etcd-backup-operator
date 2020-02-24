package etcdbackup

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd"
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

		if etcdInstance.ETCDv2.AreComplete() {
			// If state is terminal, there's nothing else we can do on this instance, so just skip to next one.
			if isTerminalInstaceState(instanceStatus.V2.Status) {
				continue
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting v2 backup on instance %s", etcdInstance.Name))

			err := r.performETCDv2Backup(ctx, etcdInstance.ETCDv2, instanceStatus.Name)
			if err == nil {
				// Backup was successful.
				instanceStatus.V2.LatestError = ""
				instanceStatus.V2.Status = InstanceBackupStateCompleted
			} else {
				// Backup was unsuccessful.
				instanceStatus.V2.LatestError = err.Error()
				instanceStatus.V2.Status = InstanceBackupStateFailed
			}

			instanceStatus.V2.FinishedTimestamp = v1alpha1.DeepCopyTime{
				Time: time.Now().UTC(),
			}
		} else {
			r.logger.LogCtx(ctx, "level", "info", "message", "V2 backup skipped for %s because ETCD V2 setting are not set.", etcdInstance.Name)
			instanceStatus.V2.Status = InstanceBackupStateSkipped
		}

		customObject.Status.Instances[etcdInstance.Name] = instanceStatus

		err = r.persistCustomObject(customObject)
		if err != nil {
			return "", microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status to '%s'", etcdInstance.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return BackupStateRunningV2BackupRunning, nil
	}

	// No status changes have happened within any of the instances, backup is completed.
	return BackupStateRunningV2BackupCompleted, nil
}

func (r *Resource) backupV2Attempt(ctx context.Context, etcdSettings giantnetes.ETCDv2Settings, instanceName string) error {
	encPass := os.Getenv("ENCRYPTION_PASSWORD")

	b := etcd.V2Backupper{
		Datadir: etcdSettings.DataDir,
		EncPass: encPass,
		Logger:  r.logger,
		Prefix:  key.GetPrefix(instanceName),
	}

	var err error
	version := b.Version()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Creating backup file")
	_, err = b.Create()
	if err != nil {
		return microerror.Maskf(err, "Etcd %s creation failed: %s", version, err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Encrypting backup file")
	filepath, err := b.Encrypt()
	if err != nil {
		return microerror.Maskf(err, "Etcd %s encryption failed: %s", version, err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Uploading backup file")
	_, err = r.upload(ctx, filepath)
	if err != nil {
		return microerror.Maskf(err, "Etcd %s upload failed: %s", version, err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Cleaning up")
	b.Cleanup()

	return nil
}

func (r *Resource) performETCDv2Backup(ctx context.Context, etcdSettings giantnetes.ETCDv2Settings, instanceName string) error {
	if !etcdSettings.AreComplete() {
		return microerror.Mask(microerror.New("EtcdV2 settings missing unexpectedly."))
	}

	attempts := 0

	o := func() error {
		attempts = attempts + 1
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for %s", attempts, instanceName))

		err := r.backupV2Attempt(ctx, etcdSettings, instanceName)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Backup attempt #%d failed for %s. Latest error was: %s", attempts, instanceName, err))
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for %s was successful", attempts, instanceName))

		return nil
	}
	b := backoff.NewMaxRetries(uint64(AllowedBackupAttempts), 20*time.Second)

	err := backoff.Retry(o, b)
	if err != nil {
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("All backup attempts failed for %s. Latest error was: %s", instanceName, err))
		return err
	}

	return nil
}

func (r *Resource) upload(ctx context.Context, filepath string) (int64, error) {
	// TODO to be implemented in a future PR
	return 0, nil
}

package etcdbackup

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd"
	"github.com/giantswarm/etcd-backup-operator/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/internal/state"
)

func (r *Resource) backupRunningV3BackupRunningTransition(ctx context.Context, obj interface{}, currentState state.State) (state.State, error) {
	doneSomething, err := r.runBackupOnAllInstances(ctx, obj, r.doV3Backup)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if doneSomething {
		return backupStateRunningV3BackupRunning, nil
	}

	// No work has been done in any of the instances, backup is completed.
	return backupStateRunningV3BackupCompleted, nil
}

func (r *Resource) doV3Backup(ctx context.Context, etcdInstance giantnetes.ETCDInstance, instanceStatus *v1alpha1.ETCDInstanceBackupStatusIndex) bool {
	etcdSettings := etcdInstance.ETCDv3

	if etcdSettings.AreComplete() {
		// If state is terminal, there's nothing else we can do on this instance, so just skip to next one.
		if isTerminalInstaceState(instanceStatus.V3.Status) {
			return false
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting v3 backup on instance %s", instanceStatus.Name))

		backupper := etcd.V3Backup{
			CACert:    etcdSettings.CaCert,
			Cert:      etcdSettings.Cert,
			EncPass:   os.Getenv("ENCRYPTION_PASSWORD"),
			Endpoints: etcdSettings.Endpoints,
			Logger:    r.logger,
			Key:       etcdSettings.Key,
			Prefix:    key.FilenamePrefix(instanceStatus.Name),
		}

		err := r.performBackup(ctx, backupper, instanceStatus.Name)
		if err == nil {
			// Backup was successful.
			instanceStatus.V3.LatestError = ""
			instanceStatus.V3.Status = instanceBackupStateCompleted
		} else {
			// Backup was unsuccessful.
			instanceStatus.V3.LatestError = err.Error()
			instanceStatus.V3.Status = instanceBackupStateFailed
		}

		instanceStatus.V3.FinishedTimestamp = v1alpha1.DeepCopyTime{
			Time: time.Now().UTC(),
		}
	} else {
		r.logger.LogCtx(ctx, "level", "info", "message", "V2 backup skipped for %s because ETCD V2 setting are not set.", instanceStatus.Name)
		instanceStatus.V2.Status = instanceBackupStateSkipped
	}

	return true
}

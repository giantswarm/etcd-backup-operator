package etcdbackup

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/etcd-backup-operator/v2/pkg/etcd"
	"github.com/giantswarm/etcd-backup-operator/v2/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/v2/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/v2/service/controller/resource/etcdbackup/internal/state"
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
	// If state is terminal, there's nothing else we can do on this instance, so just skip to next one.
	if isTerminalInstaceState(instanceStatus.V3.Status) {
		return false
	}

	if instanceStatus.V3.StartedTimestamp.Time.IsZero() {
		instanceStatus.V3.StartedTimestamp.Time = time.Now().UTC()
	}

	etcdSettings := etcdInstance.ETCDv3

	if etcdSettings.AreComplete() {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting v3 backup on instance %s", instanceStatus.Name))

		backupper := etcd.NewV3Backup(etcdSettings.CaCert, etcdSettings.Cert, r.encryptionPwd, etcdSettings.Endpoints, r.logger, etcdSettings.Key, key.FilenamePrefix(r.installation, instanceStatus.Name))

		backupAttemptResult, err := r.performBackup(ctx, backupper, instanceStatus.Name)
		if err == nil {
			// Backup was successful.
			instanceStatus.V3.LatestError = ""
			instanceStatus.V3.Status = instanceBackupStateCompleted
			instanceStatus.V3.CreationTime = backupAttemptResult.CreationTimeMeasurement
			instanceStatus.V3.EncryptionTime = backupAttemptResult.EncryptionTimeMeasurement
			instanceStatus.V3.UploadTime = backupAttemptResult.UploadTimeMeasurement
			instanceStatus.V3.BackupFileSize = backupAttemptResult.BackupSizeMeasurement
			instanceStatus.V3.Filename = backupAttemptResult.Filename
		} else {
			// Backup was unsuccessful.
			instanceStatus.V3.LatestError = err.Error()
			instanceStatus.V3.Status = instanceBackupStateFailed
		}
	} else {
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("V3 backup skipped for %s because ETCD V2 setting are not set.", instanceStatus.Name))
		instanceStatus.V3.Status = instanceBackupStateSkipped
	}

	instanceStatus.V3.FinishedTimestamp = metav1.Time{Time: time.Now().UTC()}

	return true
}

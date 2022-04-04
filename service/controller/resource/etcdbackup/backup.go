package etcdbackup

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/v3/pkg/etcd"
	"github.com/giantswarm/etcd-backup-operator/v3/pkg/etcd/metrics"
)

func (r *Resource) performBackup(ctx context.Context, backupper etcd.Backupper, instanceName string) (*metrics.BackupAttemptResult, error) {
	attempts := 0
	var err error
	var latestMetrics *metrics.BackupAttemptResult

	o := func() error {
		attempts = attempts + 1
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for %s", attempts, instanceName))

		latestMetrics, err = r.backupAttempt(ctx, backupper)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Backup attempt #%d failed for %s. Latest error was: %s", attempts, instanceName, err))
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for %s was successful", attempts, instanceName))

		return nil
	}
	b := backoff.NewMaxRetries(uint64(maxBackupAttempts), 20*time.Second)

	err = backoff.Retry(o, b)
	if err != nil {
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("All backup attempts failed for %s. Latest error was: %s", instanceName, err))
		return latestMetrics, err
	}

	return latestMetrics, nil
}

func (r *Resource) backupAttempt(ctx context.Context, b etcd.Backupper) (*metrics.BackupAttemptResult, error) {
	var err error
	version := b.Version()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Creating backup file")
	start := time.Now()
	_, err = b.Create()
	if err != nil {
		return metrics.NewFailedBackupAttemptResult(), microerror.Maskf(executionFailedError, "etcd %#q creation failed with error %#q", version, err)
	}
	creationTime := time.Since(start).Milliseconds()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Encrypting backup file")
	start = time.Now()
	path, err := b.Encrypt()
	if err != nil {
		return metrics.NewFailedBackupAttemptResult(), microerror.Maskf(executionFailedError, "etcd %#q encryption failed with error %#q", version, err)
	}
	encryptionTime := time.Since(start).Milliseconds()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Uploading backup file")
	start = time.Now()
	backupSize, err := r.uploader.Upload(path)
	if err != nil {
		return metrics.NewFailedBackupAttemptResult(), microerror.Maskf(executionFailedError, "etcd %#q upload failed with error %#q", version, err)
	}
	uploadTime := time.Since(start).Milliseconds()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Cleaning up")
	b.Cleanup()

	return metrics.NewSuccessfulBackupAttemptResult(backupSize, creationTime, encryptionTime, uploadTime, filepath.Base(path)), nil
}

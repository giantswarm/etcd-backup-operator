package etcdbackup

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd"
)

func (r *Resource) performBackup(ctx context.Context, backupper etcd.Backupper, instanceName string) error {
	attempts := 0

	o := func() error {
		attempts = attempts + 1
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for %s", attempts, instanceName))

		err := r.backupAttempt(ctx, backupper)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Backup attempt #%d failed for %s. Latest error was: %s", attempts, instanceName, err))
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for %s was successful", attempts, instanceName))

		return nil
	}
	b := backoff.NewMaxRetries(uint64(maxBackupAttempts), 20*time.Second)

	err := backoff.Retry(o, b)
	if err != nil {
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("All backup attempts failed for %s. Latest error was: %s", instanceName, err))
		return err
	}

	return nil
}

func (r *Resource) backupAttempt(ctx context.Context, b etcd.Backupper) error {
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
	_, err = r.uploader.Upload(filepath)
	if err != nil {
		return microerror.Maskf(err, "Etcd %s upload failed: %s", version, err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Cleaning up")
	b.Cleanup()

	return nil
}

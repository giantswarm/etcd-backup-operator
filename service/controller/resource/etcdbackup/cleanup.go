package etcdbackup

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	"github.com/giantswarm/microerror"
)

func (r *Resource) cleanup(ctx context.Context, etcdBackup v1alpha1.ETCDBackup) error {
	return r.cleanupByTimestamp(ctx, etcdBackup.Status.FinishedTimestamp.Time, crKeepTimeoutSeconds, etcdBackup)
}

func (r *Resource) cleanupSkippedCR(ctx context.Context, etcdBackup v1alpha1.ETCDBackup) error {
	return r.cleanupByTimestamp(ctx, etcdBackup.CreationTimestamp.Time, crSkippedKeepTimeoutSeconds, etcdBackup)
}

func (r *Resource) cleanupByTimestamp(ctx context.Context, timestamp time.Time, timeoutSeconds int64, etcdBackup v1alpha1.ETCDBackup) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Looking for ETCDBackup resources older than %d seconds", timeoutSeconds))
	diff := time.Now().UTC().Sub(timestamp).Seconds()
	if diff > float64(timeoutSeconds) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Deleting (state %s, %f seconds old)", etcdBackup.Status.Status, diff))
		err := r.k8sClient.CtrlClient().Delete(ctx, &etcdBackup)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Not due for deletion (state %s, %f seconds old)", etcdBackup.Status.Status, diff))
	}

	return nil
}

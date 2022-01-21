package etcdbackup

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
)

func (r *Resource) cleanup(ctx context.Context, etcdBackup v1alpha1.ETCDBackup) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Looking for completed ETCDBackup resources older than %d seconds", crKeepTimeoutSeconds))
	diff := time.Now().UTC().Sub(etcdBackup.Status.FinishedTimestamp.Time).Seconds()
	if diff > crKeepTimeoutSeconds {
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

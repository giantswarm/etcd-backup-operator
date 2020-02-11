package etcdbackup

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) cleanup(ctx context.Context, etcdBackup v1alpha1.ETCDBackup) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Looking for completed ETCDBackup resources older than %d seconds", CRKeepTimeoutSeconds))
	diff := time.Now().UTC().Sub(etcdBackup.Status.FinishedTimestamp.Time).Seconds()
	if diff > CRKeepTimeoutSeconds {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Deleting (state %s, %f seconds old)", etcdBackup.Status.Status, diff))
		err := r.k8sClient.G8sClient().BackupV1alpha1().ETCDBackups().Delete(etcdBackup.Name, &v1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Not due for deletion (state %s, %f seconds old)", etcdBackup.Status.Status, diff))
	}

	return nil
}

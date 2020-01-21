package etcdbackup

import (
	"context"
	"fmt"
	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// Called when Status = "Completed" or "Failed".
//   - FinishedTimestamp older than threshold?
//   yes)
//     - Delete CR.
//   no)
//     - Exit reconciliation.
func (r *Resource) cleanup(ctx context.Context, etcdBackup v1alpha1.ETCDBackup) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Looking for completed ETCDBackup resources older than %d seconds", key.CRKeepTimeoutSeconds))
	diff := time.Now().UTC().Sub(etcdBackup.Status.FinishedTimestamp.Time).Seconds()
	if diff > key.CRKeepTimeoutSeconds {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Deleting ETCDBackup %s is (state %s, %f seconds old)", etcdBackup.Name, etcdBackup.Status.Status, diff))
		err := r.K8sClient.G8sClient().BackupV1alpha1().ETCDBackups().Delete(etcdBackup.Name, &v1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ETCDBackup %s is not due for deletion (state %s, %f seconds old)", etcdBackup.Name, etcdBackup.Status.Status, diff))
	}

	return nil
}

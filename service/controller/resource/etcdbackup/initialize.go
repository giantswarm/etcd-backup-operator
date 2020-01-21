package etcdbackup

import (
	"context"
	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/microerror"
	"time"
)

// Called when Status = "" || Status = "Pending".
//   - sets StartedTimestamp to now().
//   - sets status to Running.
func (r *Resource) InitializeBackup(ctx context.Context, backup v1alpha1.ETCDBackup) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "Initializing global Status")
	backup.Status.StartedTimestamp = v1alpha1.DeepCopyTime{
		Time: time.Now().UTC(),
	}
	backup.Status.Status = key.StatusRunning
	backup.Status.Instances = []v1alpha1.ETCDInstanceBackupStatus{}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Persisting global Status")

	_, err := r.K8sClient.G8sClient().BackupV1alpha1().ETCDBackups().UpdateStatus(&backup)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Initialized global Status")

	return nil
}

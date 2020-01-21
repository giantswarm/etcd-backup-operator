package etcdbackup

import (
	"context"
	"fmt"
	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource"
	"time"
)

func (r *Resource) updateInstanceStatus(ctx context.Context, etcdBackup *v1alpha1.ETCDBackup, updatedStatus v1alpha1.ETCDInstanceBackupStatus) {
	for idx, status := range etcdBackup.Status.Instances {
		if status.Name == updatedStatus.Name {
			etcdBackup.Status.Instances[idx] = updatedStatus
			return
		}
	}

	etcdBackup.Status.Instances = append(etcdBackup.Status.Instances, updatedStatus)
}

func (r *Resource) findOrInitializeStatus(ctx context.Context, etcdBackup v1alpha1.ETCDBackup, instance resource.ETCDInstance) v1alpha1.ETCDInstanceBackupStatus {
	for _, status := range etcdBackup.Status.Instances {
		if status.Name == instance.Name {
			return status
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Initializing new ETCDInstanceBackupStatus for %s", instance.Name))

	newStatus := v1alpha1.ETCDInstanceBackupStatus{
		Name:     instance.Name,
		Status:   key.StatusPending,
		Attempts: 0,
		StartedTimestamp: v1alpha1.DeepCopyTime{
			Time: time.Now().UTC(),
		},
	}

	return newStatus
}

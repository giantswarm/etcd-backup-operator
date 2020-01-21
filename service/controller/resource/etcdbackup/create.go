package etcdbackup

import (
	"context"
	"fmt"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	etcdBackup, err := key.ToETCDBackup(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	globalStatus := etcdBackup.Status.Status

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("The global Status of ETCDBackup %s is %s", etcdBackup.Name, globalStatus))
	switch globalStatus {
	case key.StatusEmpty:
		fallthrough
	case key.StatusPending:
		err := r.InitializeBackup(ctx, etcdBackup)
		if err != nil {
			return microerror.Mask(err)
		}
	case key.StatusRunning:
		err := r.executeBackups(ctx, etcdBackup)
		if err != nil {
			if err != nil {
				return microerror.Mask(err)
			}
		}
	case key.StatusCompleted:
		fallthrough
	case key.StatusFailed:
		err := r.cleanup(ctx, etcdBackup)
		if err != nil {
			if err != nil {
				return microerror.Mask(err)
			}
		}
	default:
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Unexpected status %s for ETCDBackup %s", globalStatus, etcdBackup.Name))
	}

	return nil
}

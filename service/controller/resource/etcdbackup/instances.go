package etcdbackup

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/etcd-backup-operator/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
)

func (r *Resource) runBackupOnAllInstances(ctx context.Context, obj interface{}, handler func(context.Context, giantnetes.ETCDInstance, *v1alpha1.ETCDInstanceBackupStatusIndex) bool) (bool, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return false, microerror.Mask(err)
	}

	utils, err := giantnetes.NewUtils(r.logger, r.k8sClient)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// Control plane.
	instances := []giantnetes.ETCDInstance{
		{
			Name:   key.ControlPlane,
			ETCDv2: r.etcdV2Settings,
			ETCDv3: r.etcdV3Settings,
		},
	}

	if customObject.Spec.GuestBackup {
		// Tenant clusters.
		guestInstances, err := utils.GetTenantClusters(ctx, customObject)
		if err != nil {
			return false, microerror.Mask(err)
		}
		instances = append(instances, guestInstances...)
	}

	if len(customObject.Status.Instances) == 0 {
		customObject.Status.Instances = make(map[string]v1alpha1.ETCDInstanceBackupStatusIndex)
	}

	for _, etcdInstance := range instances {
		instanceStatus := r.findOrInitializeInstanceStatus(ctx, customObject, etcdInstance)

		doneSomething := handler(ctx, etcdInstance, &instanceStatus)

		if doneSomething {
			customObject.Status.Instances[etcdInstance.Name] = instanceStatus

			err = r.persistCustomObject(customObject)
			if err != nil {
				return false, microerror.Mask(err)
			}
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set resource status for instance '%s'", etcdInstance.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return true, nil
		}
	}

	// No status changes have happened within any of the instances, backup is completed.
	return false, nil
}

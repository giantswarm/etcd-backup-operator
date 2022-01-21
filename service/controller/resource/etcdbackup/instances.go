package etcdbackup

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/etcd-backup-operator/v2/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/v2/service/controller/key"
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

	var instances []giantnetes.ETCDInstance
	if len(customObject.Spec.ClusterNames) > 0 {
		r.logger.LogCtx(ctx, "level", "error", "message", "CR contains explicit list of cluster names")

		// User specified a list of cluster IDs to be backed up.
		// Load workload clusters.
		guestInstances, err := utils.GetTenantClusters(ctx, customObject)
		if err != nil {
			return false, microerror.Mask(err)
		}

		for _, id := range customObject.Spec.ClusterNames {
			if id == key.ManagementCluster {
				instances = append(instances, giantnetes.ETCDInstance{
					Name:   key.ManagementCluster,
					ETCDv2: r.etcdV2Settings,
					ETCDv3: r.etcdV3Settings,
				},
				)
			} else {
				found := false
				for _, candidate := range guestInstances {
					if candidate.Name == id {
						instances = append(instances, candidate)
						found = true
						break
					}
				}

				if !found {
					r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("cluster %q was not found", id))
				}
			}
		}
	} else {
		r.logger.LogCtx(ctx, "level", "error", "message", "CR does not contain explicit list of cluster names")
		// Control plane.
		instances = []giantnetes.ETCDInstance{
			{
				Name:   key.ManagementCluster,
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
	}

	if len(customObject.Status.Instances) == 0 {
		customObject.Status.Instances = make(map[string]v1alpha1.ETCDInstanceBackupStatusIndex)
	}

	for _, etcdInstance := range instances {
		instanceStatus := r.findOrInitializeInstanceStatus(ctx, customObject, etcdInstance)

		doneSomething := handler(ctx, etcdInstance, &instanceStatus)

		if doneSomething {
			customObject.Status.Instances[etcdInstance.Name] = instanceStatus

			err = r.persistCustomObjectStatus(ctx, customObject)
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

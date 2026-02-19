package etcdbackup

import (
	"context"
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/giantswarm/apiextensions-backup/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/etcd-backup-operator/v5/pkg/giantnetes"
	"github.com/giantswarm/etcd-backup-operator/v5/service/controller/key"
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

	if len(customObject.Status.Instances) == 0 {
		customObject.Status.Instances = make(map[string]v1alpha1.ETCDInstanceBackupStatusIndex)
	}

	var instances []giantnetes.ETCDInstance
	if len(customObject.Spec.ClusterNames) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "CR contains explicit list of cluster names")

		// User specified a list of cluster IDs to be backed up.
		// Load workload clusters.
		guestInstances, err := utils.GetTenantClusters(ctx)
		if err != nil {
			return false, microerror.Mask(err)
		}

		for _, id := range customObject.Spec.ClusterNames {
			if id == key.ManagementCluster {
				instances = append(instances, giantnetes.ETCDInstance{
					Name:   key.ManagementCluster,
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
					instanceStatus := r.findOrInitializeInstanceStatus(ctx, customObject, id)
					instanceStatus.Error = "No cluster found with such name or unable to initialize etcd client due to missing data. Please check etcd-backup-operator logs for more details."
					instanceStatus.V3 = nil
					customObject.Status.Instances[id] = instanceStatus
					err = r.persistCustomObjectStatus(ctx, customObject)
					if err != nil {
						return false, microerror.Mask(err)
					}
				}
			}
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "CR does not contain explicit list of cluster names")
		// Control plane.
		if !r.skipManagementClusterBackup {
			cp := giantnetes.ETCDInstance{
				Name:   key.ManagementCluster,
				ETCDv3: r.etcdV3Settings,
			}
			instances = append(instances, cp)
		}

		if customObject.Spec.GuestBackup {
			// Tenant clusters.
			guestInstances, err := utils.GetTenantClusters(ctx)
			if err != nil {
				return false, microerror.Mask(err)
			}

			clustersToIncludeRegex := regexp2.MustCompile(customObject.Spec.ClustersRegex, 0)
			clustersToExcludeRegex := regexp2.MustCompile(customObject.Spec.ClustersToExcludeRegex, 0)
			for _, guestInstance := range guestInstances {
				if isMatch, _ := clustersToIncludeRegex.MatchString(guestInstance.Name); !isMatch {
					continue
				}
				if isMatch, _ := clustersToExcludeRegex.MatchString(guestInstance.Name); isMatch {
					continue
				}

				instances = append(instances, guestInstance)
			}
		}
	}

	for _, etcdInstance := range instances {
		instanceStatus := r.findOrInitializeInstanceStatus(ctx, customObject, etcdInstance.Name)

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

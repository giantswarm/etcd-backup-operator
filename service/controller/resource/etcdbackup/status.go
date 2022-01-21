package etcdbackup

import (
	"context"
	"fmt"

	backupv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Resource) getGlobalStatus(customObject backupv1alpha1.ETCDBackup) (string, error) {
	return customObject.Status.Status, nil
}

func (r *Resource) setGlobalStatus(ctx context.Context, customObject backupv1alpha1.ETCDBackup, updatedStatus string) error {
	// Get error from API before updating it.
	obj := backupv1alpha1.ETCDBackup{}
	err := r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: customObject.Name, Namespace: customObject.Namespace}, &obj)
	if err != nil {
		return microerror.Mask(err)
	}

	obj.Status.Status = updatedStatus

	return r.persistCustomObjectStatus(ctx, obj)
}

func (r *Resource) findOrInitializeInstanceStatus(ctx context.Context, etcdBackup backupv1alpha1.ETCDBackup, instanceName string) backupv1alpha1.ETCDInstanceBackupStatusIndex {
	status, found := etcdBackup.Status.Instances[instanceName]
	if found {
		return status
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Initializing new ETCDInstanceBackupStatus for %s", instanceName))

	newStatus := backupv1alpha1.ETCDInstanceBackupStatusIndex{
		Name: instanceName,
	}

	return newStatus
}

func isTerminalInstaceState(state string) bool {
	return state == instanceBackupStateCompleted || state == instanceBackupStateFailed || state == instanceBackupStateSkipped
}

func (r *Resource) persistCustomObjectStatus(ctx context.Context, customObject backupv1alpha1.ETCDBackup) error {
	// Get error from API before updating it.
	obj := backupv1alpha1.ETCDBackup{}
	err := r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: customObject.Name, Namespace: customObject.Namespace}, &obj)
	if err != nil {
		return microerror.Mask(err)
	}

	obj.Status = customObject.Status

	err = r.k8sClient.CtrlClient().Status().Update(ctx, &obj)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

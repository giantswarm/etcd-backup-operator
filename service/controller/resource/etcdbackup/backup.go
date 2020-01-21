package etcdbackup

import (
	"context"
	"fmt"
	"github.com/giantswarm/apiextensions/pkg/apis/backup/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/etcd-backup-operator/service/controller/key"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/backup"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/giantnetes"
	"github.com/giantswarm/etcd-backup/metrics"
	"github.com/giantswarm/microerror"
	"os"
	"time"
)

// Called when Status = Running.
//   - Iterates over ETCD instances:
//     a) When instance status is "Pending":
//       - sets StartedTimestamp.
//       - sets Status "Running".
//     b) When instance status is "Running":
//       - Attempt a backup.
//       if Successful:
//         - Set status to Completed.
//         - Set FinishedTimestamp.
//       if failed:
//         - Set status to Failed.
//         - Set FinishedTimestamp.
//     c) When instance state is either "Completed" or "Failed":
//       - Continue to next iteration.
//   - Any backup attempt made?
//   yes)
//     - Exit reconciliation.
//   no)
//     - Any failed status among Instances?
//     yes)
//       - Set status to Failed.
//     no)
//       - Set status to Completed.
//     - Set FinishedTimestamp.
func (r *Resource) executeBackups(ctx context.Context, etcdBackup v1alpha1.ETCDBackup) error {
	anyWipInstances := false
	anyFailed := false

	utils := giantnetes.NewUtils(r.logger, r.K8sClient)

	// control plane
	instances := []resource.ETCDInstance{
		{
			Name:   key.ControlPlane,
			ETCDv2: r.ETCDv2Settings,
			ETCDv3: r.ETCDv3Settings,
		},
	}

	if etcdBackup.Spec.GuestBackup {
		// tenant clusters
		guestInstances, err := utils.GetTenantClusters(ctx, etcdBackup)
		if err != nil {
			return microerror.Mask(err)
		}
		instances = append(instances, guestInstances...)
	}

	for _, etcdinstance := range instances {
		instanceStatus := r.findOrInitializeStatus(ctx, etcdBackup, etcdinstance)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting working on instanceStatus %s (status %s)", instanceStatus.Name, instanceStatus.Status))
		switch instanceStatus.Status {
		case key.StatusPending:
			anyWipInstances = true
			instanceStatus.StartedTimestamp = v1alpha1.DeepCopyTime{
				Time: time.Now().UTC(),
			}
			instanceStatus.Status = key.StatusRunning
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Set status %s for instanceStatus %s", instanceStatus.Status, instanceStatus.Name))
		case key.StatusRunning:
			anyWipInstances = true
			if instanceStatus.Attempts < key.AllowedBackupAttempts {
				var latestMetrics *metrics.BackupMetrics

				o := func() error {
					instanceStatus.Attempts = instanceStatus.Attempts + 1
					r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for instanceStatus %s", instanceStatus.Attempts, instanceStatus.Name))

					err, backupMetrics := r.backupSingleInstance(ctx, etcdBackup, etcdinstance)
					latestMetrics = backupMetrics

					if err != nil {
						return microerror.Mask(err)
					}

					r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for instanceStatus %s was successful", instanceStatus.Attempts, instanceStatus.Name))
					instanceStatus.Status = key.StatusCompleted
					instanceStatus.FinishedTimestamp = v1alpha1.DeepCopyTime{
						Time: time.Now().UTC(),
					}
					// clear error message because we had a success so it's not useful anymore.
					instanceStatus.LatestError = ""

					return nil
				}

				b := backoff.NewMaxRetries(uint64(key.AllowedBackupAttempts-instanceStatus.Attempts), 20*time.Second)

				err := backoff.Retry(o, b)

				// success or failure, I send the backup metrics
				if latestMetrics != nil {
					r.ETCDBackupMetrics.Update(instanceStatus.Name, latestMetrics)
				}

				if err != nil {
					r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempt number %d for instanceStatus %s failed with the following error: %s", instanceStatus.Attempts, instanceStatus.Name, err))
					instanceStatus.LatestError = err.Error()
				}
			} else {
				instanceStatus.Status = key.StatusFailed
				instanceStatus.FinishedTimestamp = v1alpha1.DeepCopyTime{
					Time: time.Now().UTC(),
				}

				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("All %d attempts are failed for instanceStatus %s. Backup failed.", key.AllowedBackupAttempts, instanceStatus.Name))
			}
		case key.StatusCompleted:
			// nothing to do
		case key.StatusFailed:
			// nothing to do
			anyFailed = true
		}

		r.updateInstanceStatus(ctx, &etcdBackup, instanceStatus)
	}

	// Any backup attempt made?
	if anyWipInstances {
		r.logger.LogCtx(ctx, "level", "debug", "message", "No instances with WIP backups.")
	} else {
		if anyFailed {
			r.logger.LogCtx(ctx, "level", "error", "message", "At least one backup failed.")
			etcdBackup.Status.Status = key.StatusFailed
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "All backups completed successfully.")
			etcdBackup.Status.Status = key.StatusCompleted
		}
		etcdBackup.Status.FinishedTimestamp = v1alpha1.DeepCopyTime{
			Time: time.Now().UTC(),
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Persisting Status for %s.", etcdBackup.Name))

	_, err := r.K8sClient.G8sClient().BackupV1alpha1().ETCDBackups().UpdateStatus(&etcdBackup)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Status persisted for %s.", etcdBackup.Name))

	return nil
}

func (r *Resource) backupSingleInstance(ctx context.Context, etcdBackup v1alpha1.ETCDBackup, instance resource.ETCDInstance) (error, *metrics.BackupMetrics) {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Starting backup for instance %s.", instance.Name))

	encPass := os.Getenv("ENCRYPTION_PASSWORD")

	if len(instance.ETCDv2.DataDir) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempting V2 backup for instance %s.", instance.Name))

		v2 := backup.EtcdBackupV2{
			Datadir: instance.ETCDv2.DataDir,
			EncPass: encPass,
			Logger:  r.logger,
			Prefix:  key.GetPrefix(instance.Name),
		}

		err, backupMetrics := r.performV2orV3Backup(ctx, &v2)
		if err != nil {
			return microerror.Mask(err), backupMetrics
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("V2 backup successful for instance %s.", instance.Name))
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Attempting V3 backup for instance %s.", instance.Name))

	v3 := backup.EtcdBackupV3{
		CACert:    instance.ETCDv3.CaCert,
		Cert:      instance.ETCDv3.Cert,
		EncPass:   encPass,
		Endpoints: instance.ETCDv3.Endpoints,
		Logger:    r.logger,
		Key:       instance.ETCDv3.Key,
		Prefix:    key.GetPrefix(instance.Name),
	}

	err, backupMetrics := r.performV2orV3Backup(ctx, &v3)
	if err != nil {
		return microerror.Mask(err), backupMetrics
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("V3 backup successful for instance %s.", instance.Name))

	return nil, backupMetrics
}

func (r *Resource) performV2orV3Backup(ctx context.Context, b backup.BackupInterface) (error, *metrics.BackupMetrics) {
	var err error

	version := b.Version()

	start := time.Now()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Creating backup file")

	_, err = b.Create()
	if err != nil {
		return microerror.Maskf(err, "Etcd %s creation failed: %s", version, err), metrics.NewFailureMetrics()
	}

	creationTime := time.Since(start).Milliseconds()

	start = time.Now()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Encrypting backup file")

	filepath, err := b.Encrypt()
	if err != nil {
		return microerror.Maskf(err, "Etcd %s encryption failed: %s", version, err), metrics.NewFailureMetrics()
	}

	encryptionTime := time.Since(start).Milliseconds()
	start = time.Now()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Uploading backup file")

	size, err := r.upload(ctx, filepath)
	if err != nil {
		return microerror.Maskf(err, "Etcd %s upload failed: %s", version, err), metrics.NewFailureMetrics()
	}

	uploadTime := time.Since(start).Milliseconds()

	r.logger.LogCtx(ctx, "level", "debug", "message", "Cleaning up")

	b.Cleanup()

	return nil, metrics.NewSuccessfulBackupMetrics(size, creationTime, encryptionTime, uploadTime)
}

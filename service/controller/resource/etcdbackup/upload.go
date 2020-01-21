package etcdbackup

import (
	"context"
	"fmt"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/storage"
	"github.com/giantswarm/microerror"
)

func (r *Resource) upload(ctx context.Context, filepath string) (int64, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Upload strategy: S3"))

	// Configure the Storage Interface to use for storing this backup
	storageInterface := storage.NewS3(r.S3Config.Bucket, r.S3Config.Region)

	// Upload
	size, err := storageInterface.Upload(filepath)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Upload completed successfully")
	return size, nil
}

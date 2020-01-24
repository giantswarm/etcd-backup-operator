package etcdbackup

import (
	"context"
	"fmt"
	"github.com/giantswarm/microerror"
)

func (r *Resource) upload(ctx context.Context, filepath string) (int64, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Upload strategy: S3Uploader"))

	// Upload
	size, err := r.Uploader.Upload(filepath)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Upload completed successfully")
	return size, nil
}

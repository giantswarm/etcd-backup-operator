package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"os"
	"path/filepath"
)

type S3Config struct {
	AccessKeyID     string
	Bucket          string
	Region          string
	SecretAccessKey string
}

type S3Upload struct {
	accessKeyID     string
	bucket          string
	region          string
	secretAccessKey string
}

func NewS3Upload(config S3Config) (*S3Upload, error) {
	if config.AccessKeyID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AccessKeyID must be defined", config)
	}
	if config.Bucket == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Bucket must be defined", config)
	}
	if config.Region == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Region must be defined", config)

	}
	if config.SecretAccessKey == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.SecretAccessKey must be defined", config)
	}
	return &S3Upload{
		accessKeyID:     config.AccessKeyID,
		bucket:          config.Bucket,
		region:          config.Region,
		secretAccessKey: config.SecretAccessKey,
	}, nil
}

func (upload S3Upload) Upload(fpath string) (int64, error) {
	// Login to AWS S3Upload
	creds := credentials.NewStaticCredentials(upload.accessKeyID, upload.secretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      &upload.region,
	})
	if err != nil {
		return -1, microerror.Mask(err)
	}

	svc := s3.New(sess, &aws.Config{
		Region: &upload.region,
	})

	// Upload.
	file, err := os.Open(fpath)
	if err != nil {
		return -1, microerror.Mask(err)
	}
	defer file.Close()

	// Get file size.
	fileInfo, err := file.Stat()
	if err != nil {
		return -1, microerror.Mask(err)
	}
	size := fileInfo.Size()

	// Get filename without path.
	path := filepath.Base(fileInfo.Name())

	params := &s3.PutObjectInput{
		Bucket:        aws.String(upload.bucket),
		Key:           aws.String(path),
		Body:          file,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/octet-stream"),
	}

	// Put object to S3Upload.
	_, err = svc.PutObject(params)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	return size, nil
}

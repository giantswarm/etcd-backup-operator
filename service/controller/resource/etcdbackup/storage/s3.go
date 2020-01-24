package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"os"
	"path/filepath"
)

const (
	EnvAWSAccessKeyID     = "AWS_ACCESS_KEY_ID"
	EnvAWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

type S3 struct {
	Bucket string
	Region string
}

func NewS3(bucket string, region string) *S3 {
	return &S3{
		Bucket: bucket,
		Region: region,
	}
}

func (config *S3) Upload(fpath string) (int64, error) {
	// requires the following env variables to be set:
	// - AWS_ACCESS_KEY_ID
	// - AWS_SECRET_ACCESS_KEY
	required := []string{EnvAWSAccessKeyID, EnvAWSSecretAccessKey}
	for _, varname := range required {
		_, defined := os.LookupEnv(varname)
		if !defined {
			return -1, microerror.Mask(newMissingRequiredEnvVariableError(varname))
		}
	}

	// Login to AWS S3
	sess, err := session.NewSession()
	if err != nil {
		return -1, microerror.Mask(err)
	}

	svc := s3.New(sess, &aws.Config{
		Region: &config.Region,
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
		Bucket:        aws.String(config.Bucket),
		Key:           aws.String(path),
		Body:          file,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/octet-stream"),
	}

	// Put object to S3.
	_, err = svc.PutObject(params)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	return size, nil
}

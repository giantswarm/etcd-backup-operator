package service

type S3Uploader struct {
	Bucket         string
	Region         string
	Endpoint       string
	ForcePathStyle string
}

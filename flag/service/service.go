package service

import (
	"github.com/giantswarm/operatorkit/v7/pkg/flag/service/kubernetes"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Kubernetes   kubernetes.Kubernetes
	S3           S3Uploader
	ETCDv2       ETCDv2Settings
	ETCDv3       ETCDv3Settings
	Installation string
	Sentry       Sentry
}

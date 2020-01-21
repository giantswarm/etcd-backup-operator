package service

import (
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource"
	"github.com/giantswarm/etcd-backup-operator/service/controller/resource/etcdbackup/storage"
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Kubernetes kubernetes.Kubernetes
	S3         storage.S3
	ETCDv2     resource.ETCDv2Settings
	ETCDv3     resource.ETCDv3Settings
}

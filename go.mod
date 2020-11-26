module github.com/giantswarm/etcd-backup-operator/v2

go 1.13

require (
	github.com/aws/aws-sdk-go v1.35.35
	github.com/coreos/go-semver v0.3.0
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/giantswarm/apiextensions/v2 v2.6.2
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v4 v4.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.3.4
	github.com/giantswarm/operatorkit/v2 v2.0.2
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/go-cmp v0.5.4
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/prometheus/client_golang v1.8.0
	github.com/spf13/viper v1.7.1
	github.com/ulikunitz/xz v0.5.8 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
)

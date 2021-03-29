module github.com/giantswarm/etcd-backup-operator/v2

go 1.13

require (
	github.com/aws/aws-sdk-go v1.38.7
	github.com/coreos/go-semver v0.3.0
	github.com/giantswarm/apiextensions/v2 v2.6.2
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/exporterkit v0.2.1
	github.com/giantswarm/k8sclient/v4 v4.1.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v2 v2.0.2
	github.com/google/go-cmp v0.5.5
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mholt/archiver/v3 v3.5.0
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/viper v1.7.1
	github.com/ulikunitz/xz v0.5.8 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
)

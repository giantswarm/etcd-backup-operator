module github.com/giantswarm/etcd-backup-operator/v2

go 1.16

require (
	github.com/aws/aws-sdk-go v1.42.46
	github.com/coreos/go-semver v0.3.0
	github.com/giantswarm/apiextensions-backup v0.1.0
	github.com/giantswarm/apiextensions/v3 v3.40.0
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/exporterkit v1.0.0
	github.com/giantswarm/k8sclient/v7 v7.0.1
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/operatorkit/v7 v7.0.1
	github.com/google/go-cmp v0.5.7
	github.com/mholt/archiver/v3 v3.5.1
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/viper v1.10.1
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838
	k8s.io/api v0.20.15
	k8s.io/apimachinery v0.20.15
	k8s.io/client-go v0.20.15
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)

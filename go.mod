module github.com/giantswarm/etcd-backup-operator/v2

go 1.13

require (
	github.com/aws/aws-sdk-go v1.42.34
	github.com/coreos/go-semver v0.3.0
	github.com/giantswarm/apiextensions/v3 v3.39.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/exporterkit v0.2.1
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/operatorkit/v4 v4.3.1
	github.com/google/go-cmp v0.5.6
	github.com/mholt/archiver/v3 v3.5.1
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/viper v1.10.1
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	k8s.io/apimachinery v0.18.19
	k8s.io/client-go v0.18.19
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)

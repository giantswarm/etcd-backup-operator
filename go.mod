module github.com/giantswarm/etcd-backup-operator

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/aws/aws-sdk-go v1.28.0
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/giantswarm/apiextensions v0.0.0-20200116130625-bef324aa6223
	github.com/giantswarm/backoff v0.0.0-20190913091243-4dd491125192
	github.com/giantswarm/etcd-backup v0.0.0-20200108121155-402c6bedf381
	github.com/giantswarm/exporterkit v0.0.0-20190619131829-9749deade60f
	github.com/giantswarm/k8sclient v0.0.0-20191213144452-f75fead2ae06
	github.com/giantswarm/microendpoint v0.0.0-20191121160659-e991deac2653
	github.com/giantswarm/microerror v0.0.0-20191011121515-e0ebc4ecf5a5
	github.com/giantswarm/microkit v0.0.0-20191023091504-429e22e73d3e
	github.com/giantswarm/micrologger v0.0.0-20191014091141-d866337f7393
	github.com/giantswarm/operatorkit v0.0.0-20200114125246-3ab5e82b3050
	github.com/giantswarm/to v0.0.0-20191022113953-f2078541ec95 // indirect
	github.com/giantswarm/versionbundle v0.0.0-20191206123034-be95231628ae
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pierrec/lz4 v2.4.0+incompatible // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.1.0 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.1
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	golang.org/x/sys v0.0.0-20191210023423-ac6580df4449 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20200113154838-30cae5f2fb06 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/api v0.15.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	honnef.co/go/tools v0.0.1-2019.2.3 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833 // indirect
	k8s.io/apimachinery v0.16.5-beta.0
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7

module github.com/giantswarm/etcd-backup-operator

go 1.13

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/giantswarm/apiextensions v0.0.0-20200207112243-2a2d71e5388c
	github.com/giantswarm/backoff v0.0.0-20190913091243-4dd491125192 // indirect
	github.com/giantswarm/exporterkit v0.0.0-20190619131829-9749deade60f
	github.com/giantswarm/k8sclient v0.0.0-20191213144452-f75fead2ae06
	github.com/giantswarm/microendpoint v0.0.0-20191121160659-e991deac2653
	github.com/giantswarm/microerror v0.0.0-20191011121515-e0ebc4ecf5a5
	github.com/giantswarm/microkit v0.0.0-20191023091504-429e22e73d3e
	github.com/giantswarm/micrologger v0.0.0-20191014091141-d866337f7393
	github.com/giantswarm/operatorkit v0.0.0-20200116124438-aa7ed9599161
	github.com/giantswarm/to v0.0.0-20191022113953-f2078541ec95 // indirect
	github.com/giantswarm/versionbundle v0.0.0-20191206123034-be95231628ae
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.1.0 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/spf13/viper v1.6.1
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	golang.org/x/sys v0.0.0-20191210023423-ac6580df4449 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.0.0
	k8s.io/kubernetes v1.17.2 // indirect
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
)

replace (
	gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
	k8s.io/api => k8s.io/api v0.0.0-20200131193051-d9adff57e763
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20200131201446-6910daba737d
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3-beta.0.0.20200131192631-731dcecc2054
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20200131195721-b64b0ef70370
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20200131202043-1dc23f43cc94
	k8s.io/client-go => k8s.io/client-go v0.17.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20200131203830-fe5589c708de
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20200131203557-3c6746d7c617
	k8s.io/code-generator => k8s.io/code-generator v0.17.3-beta.0.0.20200131192142-4ae19cfe9b46
	k8s.io/component-base => k8s.io/component-base v0.0.0-20200131194811-85b325a9731b
	k8s.io/cri-api => k8s.io/cri-api v0.17.3-beta.0.0.20200131204836-cb8a25f43f0e
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20200131204100-4311b557c8ce
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20200131200134-d62c64b672cc
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20200131203333-c935c9222556
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20200131202556-6b094e7591d1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20200131203102-8e9ee8fa0785
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20200131205129-9ef1401eb3ec
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20200131202828-eb1b5c1ce7fb
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20200131204342-ef4bac7ed518
	k8s.io/metrics => k8s.io/metrics v0.0.0-20200131201757-ffbb7a48f604
	k8s.io/node-api => k8s.io/node-api v0.0.0-20200131204614-47835c5f2652
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20200131200511-51b2302b2589
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20200131202323-14126e90c844
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20200131200932-3fd12213be16

)

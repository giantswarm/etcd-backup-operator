[![CircleCI](https://circleci.com/gh/giantswarm/etcd-backup-operator.svg?&style=shield&circle-token=cfd916d774b98d091010b3cfd102168b77bfc635)](https://circleci.com/gh/giantswarm/etcd-backup-operator)

# etcd-backup-operator

The `etcd-backup-operator` takes backups of ETCD instances on both the control plane and tenant clusters.

The operator is meant to be run on the CP and can perform both V2 and V3 ETCD backups (see https://www.mirantis.com/blog/everything-you-ever-wanted-to-know-about-using-etcd-with-kubernetes-v1-6-but-were-afraid-to-ask/).  

## Branches

- `master`
  - When updated, it triggers a deployment on all installations.
  
## Getting Project

Clone the git repository: https://github.com/giantswarm/etcd-backup-operator.git

### How to build

Build it using the standard `go build` command.

```
cd etcd-backup-operator
go build
```

### How to deploy

Use `opsctl` the usual way. This project uses the `app` method (`opsctl deploy ... -m app`).

## Running

Example command run:
```
export AWS_ACCESS_KEY_ID=<S3 access key ID>
export AWS_SECRET_ACCESS_KEY=<S3 secret access key>
go run -mod=vendor main.go daemon \
  --service.kubernetes.incluster="true" \
  --service.s3.bucket=<S3 bucket> \
  --service.s3.region=<S3 region> \
  --service.etcdv2.datadir="<Path of the directory where the V2 ETCD data is stored>" \
  --service.etcdv3.cacert=<Path of the ETCD CA file> \
  --service.etcdv3.cert=<Path of the ETCD Cert file> \
  --service.etcdv3.key=<Path of the ETCD Private Key file> \
  --service.etcdv3.endpoints=<URL to connect to ETCD with V3 protocol>
```

### Available flags:

#### Kubernetes connection settings:

- `--service.kubernetes.incluster`: (Optional, defaults to `false`) Whether to use the in-cluster config to authenticate with Kubernetes.
- `--service.kubernetes.address`: (Optional, defaults to `http://127.0.0.1:6443`) Address used to connect to Kubernetes. When empty in-cluster config is created.
- `--service.kubernetes.kubeconfig`: (Optional) KubeConfig used to connect to Kubernetes. When empty other settings are used.
- `--service.kubernetes.tls.cafile`: (Optional) Certificate authority file path to use to authenticate with Kubernetes.
- `--service.kubernetes.tls.crtfile`: (Optional) Certificate file path to use to authenticate with Kubernetes.
- `--service.kubernetes.tls.keyfile`: (Optional) Key file path to use to authenticate with Kubernetes.

#### S3 settings:

- `--service.s3.bucket`: (Required) AWS S3 Bucket name.
- `--service.s3.region`: (Required) AWS S3 Region name.

#### ETCD connection settings:

- `--service.etcdv2.datadir`: (Optional, see below for details) ETCD v2 Data Dir path.
- `--service.etcdv3.cert`: (Optional, see below for details) Client certificate for ETCD v3 connection
- `--service.etcdv3.cacert`: (Optional, see below for details) Client CA certificate for ETCD v3 connection
- `--service.etcdv3.key`: (Optional, see below for details) Client private key for ETCD v3 connection
- `--service.etcdv3.endpoints`: (Optional, see below for details) 

Either `service.etcdv2.datadir` or all other fields are mandatory.

You can specify all of them as well (and you'll enable both V2 and V3 backups).

#### Environment variables:

- `AWS_ACCESS_KEY_ID`: (Required) The AWS access key ID, used to upload the backup files to AWS S3. 
- `AWS_SECRET_ACCESS_KEY`: (Required) The AWS secret access key, used to upload the backup files to AWS S3.

## License

azure-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for
details.

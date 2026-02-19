[![CircleCI](https://dl.circleci.com/status-badge/img/gh/giantswarm/etcd-backup-operator.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/giantswarm/etcd-backup-operator)

# etcd-backup-operator

The `etcd-backup-operator` takes backups of ETCD instances on both the control plane and tenant clusters.

The operator is meant to be run on the management cluster and performs V3 ETCD backups.

## Branches

- `main`
  - When released, it triggers a deployment on all installations.
  
## Getting Project

Clone the Git repository: https://github.com/giantswarm/etcd-backup-operator.git

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
  --service.etcdv3.cacert=<Path of the ETCD CA file> \
  --service.etcdv3.cert=<Path of the ETCD Cert file> \
  --service.etcdv3.key=<Path of the ETCD Private Key file> \
  --service.etcdv3.endpoints=<URL to connect to ETCD with V3 protocol>
```

### Available flags:

#### Kubernetes connection settings:

- `--service.kubernetes.incluster`: (Optional, defaults to `false`) Whether to use the in-cluster config to authenticate with Kubernetes.
- `--service.kubernetes.address`: (Optional, defaults to `http://127.0.0.1:6443`) Address used to connect to Kubernetes. When empty in-cluster config is created.
- `--service.kubernetes.kubeconfig`: (Optional) KubeConfig (formatted as JSON string) used to connect to Kubernetes. When empty other settings are used.
- `--service.kubernetes.tls.cafile`: (Optional) Certificate authority file path to use to authenticate with Kubernetes.
- `--service.kubernetes.tls.crtfile`: (Optional) Certificate file path to use to authenticate with Kubernetes.
- `--service.kubernetes.tls.keyfile`: (Optional) Key file path to use to authenticate with Kubernetes.

#### S3 settings:

- `--service.s3.bucket`: (Required) AWS S3 Bucket name.
- `--service.s3.region`: (Required) AWS S3 Region name.
- `--service.s3.endpoint`: (Optional) Custom S3 endpoint URL.
- `--service.s3.force-path-style`: (Optional, defaults to `false`) Enable path-style S3 URLs.

#### IAM Roles for Service Accounts (IRSA) settings:

- `--service.enableIRSA`: (Optional, defaults to `false`) Enable IAM Roles for Service Accounts (IRSA) for S3 access instead of using static credentials.
- `--service.roleArn`: (Optional) AWS IAM Role ARN to use when IRSA is enabled.

#### AWS Authentication:

There are two ways to authenticate with AWS for S3 access:

1. **Static Credentials** (default method):
   - Set environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`

2. **IAM Roles for Service Accounts (IRSA)**:
   - Enable with `--service.enableIRSA=true`
   - Specify the role ARN with `--service.roleArn=arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME`
   - Ensure the Kubernetes service account is properly annotated with `eks.amazonaws.com/role-arn`
   - No static credentials needed when using IRSA

When IRSA is enabled, the operator will use the AWS SDK's credential chain to authenticate, which will automatically use the IAM role associated with the service account.

#### ETCD connection settings:

- `--service.etcdv3.cert`: (Required) Client certificate for ETCD v3 connection
- `--service.etcdv3.cacert`: (Required) Client CA certificate for ETCD v3 connection
- `--service.etcdv3.key`: (Required) Client private key for ETCD v3 connection
- `--service.etcdv3.endpoints`: (Required) Endpoints for ETCD v3 connection

All four ETCD v3 fields are required when management cluster backup is enabled.

#### Environment variables:

- `AWS_ACCESS_KEY_ID`: (Required) The AWS access key ID, used to upload the backup files to AWS S3. 
- `AWS_SECRET_ACCESS_KEY`: (Required) The AWS secret access key, used to upload the backup files to AWS S3.

#### Different schedules

You can schedule different cron datetimes to different clusters like it is explain here:

```yaml
schedules:
- cronjob: 0 */6 * * *
  clusters: '^(?!<cluster-id>)' # all clusters but the id defined
- cronjob: 0 3 * * * *
  clusters: '<cluster-id>' # only one cluster
```

## License

etcd-backup-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

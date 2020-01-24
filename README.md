[![CircleCI](https://circleci.com/gh/giantswarm/etcd-backup-operator.svg?&style=shield)](https://circleci.com/gh/giantswarm/etcd-backup-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/etcd-backup-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/etcd-backup-operator)

# etcd-backup-operator

Implementation Idea:

* The current cronjob creates a EtcdBackup CR which is a recipe for given backup.
* The operator reconciles on that CR and updates its Status accordingly during operation.
* Once the CR state is DONE / FINISHED and age > givenOldAgeThreshold (like 2-7 days or something) it gets deleted (by backup operator).

How to run the project locally, towards a `kind` cluster called `testing`:

```
go run -mod=vendor main.go daemon --service.kubernetes.kubeconfig="`kind get kubeconfig --name=testing`"
```

### Development

You can test locally using `kind`.

```
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
go run -mod=vendor main.go daemon \
--service.kubernetes.kubeconfig="`kind get kubeconfig`" \
--service.s3.bucket=etcd-backup \
--service.s3.region=eu-west-1 \
--service.etcdv2.datadir=<etcd storage dir for v2 backup> \
--service.etcdv3.cacert=<path to etcd client ca file> \
--service.etcdv3.cert=<path to etcd client cert file> \
--service.etcdv3.key=<path to etcd client key file>" \
--service.etcdv3.endpoints=https://<etcd endpoint>
```

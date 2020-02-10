FROM alpine:3.10

RUN apk add --no-cache ca-certificates

ADD ./etcd-backup-operator /etcd-backup-operator

ENTRYPOINT ["/etcd-backup-operator"]

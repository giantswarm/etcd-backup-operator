FROM alpine:3.10

RUN apk add --no-cache ca-certificates curl

ENV ETCD3_VER=v3.2.4
RUN \
  mkdir /usr/local/bin/etcd3 && \
  cd /tmp && \
  curl -L https://storage.googleapis.com/etcd/${ETCD3_VER}/etcd-${ETCD3_VER}-linux-amd64.tar.gz | \
  tar xz -C /usr/local/bin/etcd3 --strip-components=1

ENV ETCD2_VER=v2.3.8
RUN \
  mkdir /usr/local/bin/etcd2 && \
  cd /tmp && \
  curl -L https://storage.googleapis.com/etcd/${ETCD2_VER}/etcd-${ETCD2_VER}-linux-amd64.tar.gz | \
  tar xz -C /usr/local/bin/etcd2 --strip-components=1

ADD ./etcd-backup-operator /etcd-backup-operator

ENTRYPOINT ["/etcd-backup-operator"]

FROM alpine:3.14.1

RUN apk add --no-cache ca-certificates curl

ENV ETCD3_VER=v3.2.4
RUN \
  cd /tmp && \
  curl -L https://storage.googleapis.com/etcd/${ETCD3_VER}/etcd-${ETCD3_VER}-linux-amd64.tar.gz | \
  tar xz -C /usr/local/bin --strip-components=1

ENV KUBECTL_VERSION=v1.17.0
RUN curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /bin/kubectl && \
 chmod +x /bin/kubectl

ADD create-cr.sh /bin/create-cr.sh
RUN chmod +x /bin/create-cr.sh

ADD ./etcd-backup-operator /etcd-backup-operator

ENTRYPOINT ["/etcd-backup-operator"]

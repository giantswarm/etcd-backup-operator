FROM alpine:3.20.3

ARG TARGETOS
ARG TARGETARCH

ARG ETCD_VERSION=v3.5.14
ADD https://storage.googleapis.com/etcd/${ETCD_VERSION}/etcd-${ETCD_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz etcd.tar.gz
RUN tar xf etcd.tar.gz --directory /usr/local/bin --strip-components 1 && rm etcd.tar.gz

ARG KUBECTL_VERSION=v1.29.6
ADD --chmod=755 https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/${TARGETOS}/${TARGETARCH}/kubectl /usr/local/bin/kubectl

COPY create-cr.sh /usr/local/bin/create-cr.sh
COPY etcd-backup-operator /etcd-backup-operator

ENTRYPOINT [ "/etcd-backup-operator" ]

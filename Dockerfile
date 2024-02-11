FROM registry.ci.openshift.org/openshift/release:golang-1.20 AS builder
WORKDIR /go/src/github.com/openshift/extended-job-manager-operator
COPY . .

RUN make build

FROM registry.ci.openshift.org/openshift/origin-v4.0:base
COPY --from=builder /go/src/github.com/openshift/extended-job-manager-operator/extended-job-manager-operator /usr/bin/
# Upstream bundle and index images does not support versioning so
# we need to copy a specific version under /manifests layout directly
COPY --from=builder /go/src/github.com/openshift/extended-job-manager-operator/manifests/* /manifests/

LABEL io.k8s.display-name="OpenShift extended-job-manager Operator" \
      io.k8s.description="This is a component of OpenShift and manages the extended job manager (kueue)" \
      io.openshift.tags="openshift,extended-job-manager-operator" \
      com.redhat.delivery.appregistry=true \
      maintainer="AOS workloads team, <aos-workloads@redhat.com>"
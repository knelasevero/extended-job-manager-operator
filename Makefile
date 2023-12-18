all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/images.mk \
	targets/openshift/codegen.mk \
	targets/openshift/deps.mk \
	targets/openshift/crd-schema-gen.mk \
)

# Exclude e2e tests from unit testing
GO_TEST_PACKAGES :=./pkg/...

IMAGE_REGISTRY :=registry.svc.ci.openshift.org

CODEGEN_OUTPUT_PACKAGE :=github.com/knelasevero/extended-job-manager-operator/pkg/generated
CODEGEN_API_PACKAGE :=github.com/knelasevero/extended-job-manager-operator/pkg/apis
CODEGEN_GROUPS_VERSION :=extendedjobmanager:v1

# This will call a macro called "build-image" which will generate image specific targets based on the parameters:
# $0 - macro name
# $1 - target name
# $2 - image ref
# $3 - Dockerfile path
# $4 - context directory for image build
$(call build-image,ocp-extended-job-manager-operator,$(IMAGE_REGISTRY)/ocp/4.9:extended-job-manager-operator, ./Dockerfile.rhel7,.)

$(call verify-golang-versions,Dockerfile.rhel7)

$(call add-crd-gen,extendedjobmanager,./pkg/apis/extendedjobmanager/v1,./manifests,./manifests)

test-e2e: GO_TEST_PACKAGES :=./test/e2e
# the e2e imports pkg/cmd which has a data race in the transport library with the library-go init code
test-e2e: GO_TEST_FLAGS :=-v
test-e2e: test-unit
.PHONY: test-e2e

regen-crd:
	go build -o _output/tools/bin/controller-gen ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen
	cp manifests/extended-job-manager-operator.crd.yaml manifests/operator.openshift.io_extendedjobmanagers.yaml
	./_output/tools/bin/controller-gen crd paths=./pkg/apis/extendedjobmanager/v1/... schemapatch:manifests=./manifests output:crd:dir=./manifests
	mv manifests/operator.openshift.io_extendedjobmanagers.yaml manifests/extended-job-manager-operator.crd.yaml

generate: update-codegen-crds generate-clients
.PHONY: generate

generate-clients:
	bash ./vendor/k8s.io/code-generator/generate-groups.sh all github.com/knelasevero/extended-job-manager-operator/pkg/generated github.com/knelasevero/extended-job-manager-operator/pkg/apis extendedjobmanager:v1
.PHONY: generate-clients

clean:
	$(RM) ./extended-job-manager-operator
	$(RM) -r ./_tmp
.PHONY: clean
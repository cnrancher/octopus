# Borrowed from https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
CURR_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

all: protos generate manifests fmt vet

.PHNOY: fmt
fmt:
	go fmt ./...

.PHNOY: vet
vet:
	go vet ./...

.PHNOY: generate
generate:
	$(CURR_DIR)/hack/make-rules/controller-gen.sh \
		object:headerFile=$(CURR_DIR)/hack/boilerplate.go.txt \
		paths="$(CURR_DIR)/api/..."

.PHNOY: manifests
manifests:
	$(CURR_DIR)/hack/make-rules/controller-gen.sh \
		crd:crdVersions=v1 \
		paths="$(CURR_DIR)/api/..." \
		output:crd:dir=$(CURR_DIR)/deploy/manifests/crd
	$(CURR_DIR)/hack/make-rules/controller-gen.sh \
		webhook \
		paths="$(CURR_DIR)/api/..." \
        output:webhook:dir=$(CURR_DIR)/deploy/manifests/brain
	$(CURR_DIR)/hack/make-rules/controller-gen.sh \
		rbac:roleName=octopus-brain \
		paths="$(CURR_DIR)/pkg/brain/..." \
		output:rbac:dir=$(CURR_DIR)/deploy/manifests/brain
	$(CURR_DIR)/hack/make-rules/controller-gen.sh \
		rbac:roleName=octopus-limb \
		paths="$(CURR_DIR)/pkg/limb/..." \
		output:rbac:dir=$(CURR_DIR)/deploy/manifests/limb

.PHONY: protos
protos:
	$(CURR_DIR)/hack/make-rules/protoc.sh $(CURR_DIR)/pkg/adaptor/api/v1alpha1

.PHONY: adaptors
adaptors:
	make -f $(CURR_DIR)/adaptors/Makefile

# Borrowed from https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
CURR_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

ROOT_DIR := $(abspath $(CURR_DIR)/../..)

all: generate manifests fmt vet

.PHNOY: fmt
fmt:
	go fmt $(CURR_DIR)/...

.PHNOY: vet
vet:
	go vet $(CURR_DIR)/...

.PHNOY: generate
generate:
	$(ROOT_DIR)/hack/make-rules/controller-gen.sh \
		object:headerFile=$(ROOT_DIR)/hack/boilerplate.go.txt \
		paths="$(CURR_DIR)/api/..."

.PHNOY: manifests
manifests:
	$(ROOT_DIR)/hack/make-rules/controller-gen.sh \
		crd:crdVersions=v1 \
		paths="$(CURR_DIR)/api/..." \
		output:crd:dir=$(CURR_DIR)/deploy/manifests/crd

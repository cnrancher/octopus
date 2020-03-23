SHELL := /bin/bash

# Borrowed from https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
CURR_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

# Borrowed from https://stackoverflow.com/questions/2214575/passing-arguments-to-make-run
CMD_CNT := $(words $(MAKECMDGOALS))
FIS_CMD := $(firstword $(MAKECMDGOALS))
LST_CMD := $(lastword $(MAKECMDGOALS))
ifeq (octopus, $(FIS_CMD))
    ifeq (help, $(LST_CMD))
        RUN_ARGS := $(wordlist 2, $(shell expr $(CMD_CNT) - 1), $(MAKECMDGOALS))
    else
        RUN_ARGS := $(wordlist 2, $(CMD_CNT), $(MAKECMDGOALS))
    endif
    $(eval $(RUN_ARGS):;@:)
else ifeq (adaptor, $(FIS_CMD))
    ifeq (help, $(LST_CMD))
        RUN_ARGS := $(wordlist 2, $(shell expr $(CMD_CNT) - 1), $(MAKECMDGOALS))
    else
        RUN_ARGS := $(wordlist 2, $(CMD_CNT), $(MAKECMDGOALS))
    endif
    $(eval $(RUN_ARGS):;@:)
    ADAPTOR_MKFILE := $(CURR_DIR)/adaptors/$(word 1, $(RUN_ARGS))/Makefile
else ifneq (help, $(FIR_CMD))
    RUN_ARGS := $(wordlist 2, $(CMD_CNT), $(MAKECMDGOALS))
    $(eval $(RUN_ARGS):;@:)
    # print usage information
.PHONY: $(FIS_CMD)
$(FIS_CMD): help
	@echo "please follow the usage above !!!"
endif

.PHONY: all
all: help

.PHONY: help
help:
	# building process.
	#
	# usage:
	#   make <component {-}> stage [only]
	#
	# component:
	#   -                octopus  :  the octopus core
	#   - adaptor {adaptor-name}  :  the named adaptor
	#
	# stage:
	#   a "stage" consists of serval actions, actions follow as below:
	#     - [dev]  :  generate -> mod -> lint -> build -> test -> verify
	#     - [prd]  :                       \ = = = = = = = =  package  = = = = = = = = > e2e -> deploy
	#                                         \ -> build -> test -> containerize -> /
	#   for convenience, the name of the "action" also represents the current "stage".
	#   choosing to execute a certain "stage" will execute all actions in the previous sequence.
	#
	# actions:
	#   - generate, gen, g  :  generate deployment manifests and code implementations via `controller-gen`,
	#                          generate gPRC interfaces via `protoc`.
	#   -           mod, m  :  download code dependencies.
	#   -          lint, l  :  verify code via `golangci-lint`,
	#                          roll back to `go fmt` and `go vet` if the installation fails.
	#   -         build, b  :  compile code.
	#   -          test, t  :  run unit tests.
	#   -        verify, v  :  run integration tests.
	#   -     containerize  :  package docker image.
	#   -  package, pkg, p  :  use `dapper` to build, test and containerize.
	#   -           e2e, e  :  run e2e tests.
	#   -        deploy, d  :  push docker image.
	#   only executing the corresponding "action" of a "stage" needs the `only` suffix.
	#
	# example:
	#   -                  make octopus  :  execute `pacakge` stage for octopus.
	#   -         make octopus generate  :  execute `generate` stage for octopus.
	#   -            make adaptor dummy  :  execute `pacakge` stage for "dummy" adaptor.
	#   -       make adaptor dummy test  :  execute `test` stage for "dummy" adaptor.
	#   - make adaptor dummy build only  :  execute `build` action for "dummy" adaptor.
	@echo

.PHONY: octopus
octopus:
	@$(CURR_DIR)/hack/make-rules/octopus.sh $(RUN_ARGS)

.PHONY: adaptor
ifeq ($(ADAPTOR_MKFILE), $(wildcard $(ADAPTOR_MKFILE)))
adaptor:
	@make -se -f $(ADAPTOR_MKFILE) adaptor $(RUN_ARGS)
else
adaptor:
	@echo "does not exist '$(word 1, $(RUN_ARGS))' adaptor !!!"
endif

.PHONY: test deploy pkg

# Borrowed from https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
CURR_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

# Borrowed from https://stackoverflow.com/questions/2214575/passing-arguments-to-make-run
CMD_CNT := $(words $(MAKECMDGOALS))
FIS_CMD := $(firstword $(MAKECMDGOALS))
LST_CMD := $(lastword $(MAKECMDGOALS))
ifeq (adaptor, $(FIS_CMD))
    ifeq (help, $(LST_CMD))
        RUN_ARGS := $(wordlist 2, $(shell expr $(CMD_CNT) - 1), $(MAKECMDGOALS))
    else
        RUN_ARGS := $(wordlist 2, $(CMD_CNT), $(MAKECMDGOALS))
    endif
    $(eval $(RUN_ARGS):;@:)
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
	# Building process.
	#
	# Usage:
	#   make adaptor {adaptor-name} stage [only]
	#
	# Stage:
	#   a "stage" consists of serval actions, actions follow as below:
	#     	generate -> mod -> lint -> build -> containerize -> deploy
	#                                      \ -> test -> verify -> e2e
	#   for convenience, the name of the "action" also represents the current "stage".
	#   choosing to execute a certain "stage" will execute all actions in the previous sequence.
	#
	# Actions:
	#   -      generate, g  :  generate deployment manifests and code implementations via `controller-gen`.
	#   -           mod, m  :  download code dependencies.
	#   -          lint, l  :  verify code via `golangci-lint`,
	#                          roll back to `go fmt` and `go vet` if the installation fails.
	#   -         build, b  :  compile code.
	#   -  containerize, c  :  package docker image.
	#   -        deploy, d  :  push docker image.
	#   -          test, t  :  run unit tests.
	#   -        verify, v  :  run integration tests.
	#   -           e2e, e  :  run e2e tests.
	#   only executing the corresponding "action" of a "stage" needs the `only` suffix.
	#   integrate with dapper via `BY=dapper`.
	#
	# Example:
	#   -            make adaptor template  :  execute `build` stage for "template" adaptor.
	#   -       make adaptor template test  :  execute `test` stage for "template" adaptor.
	#   - make adaptor template build only  :  only execute `build` action for "template" adaptor, during `build` stage.
	@echo

.PHONY: adaptor
adaptor:
	@$(CURR_DIR)/hack/make-rules/adaptor.sh $(RUN_ARGS)

.PHONY: test deploy pkg

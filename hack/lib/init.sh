#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

unset CDPATH

export GO111MODULE=auto
export GOPATH=$(go env GOPATH)
export GOARCH=$(go env GOARCH)
# Set no_proxy for localhost if behind a proxy, otherwise,
# the connections to localhost in scripts will time out
export no_proxy=127.0.0.1,localhost

# The root of the octopus directory
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

source "${ROOT_DIR}/hack/lib/log.sh"
source "${ROOT_DIR}/hack/lib/controller-gen.sh"
source "${ROOT_DIR}/hack/lib/protoc.sh"
source "${ROOT_DIR}/hack/lib/find.sh"
source "${ROOT_DIR}/hack/lib/lint.sh"
source "${ROOT_DIR}/hack/lib/version.sh"
source "${ROOT_DIR}/hack/lib/docker.sh"
source "${ROOT_DIR}/hack/lib/kubectl.sh"
source "${ROOT_DIR}/hack/lib/dapper.sh"

octopus::log::install_errexit
octopus::version::get_version_vars

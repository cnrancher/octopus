#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

unset CDPATH

export GO111MODULE=auto

# The root of the octopus directory
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

OUTPUT_PATH="${ROOT_DIR}/dist"

# Set no_proxy for localhost if behind a proxy, otherwise,
# the connections to localhost in scripts will time out
export no_proxy=127.0.0.1,localhost

source "${ROOT_DIR}/hack/lib/log.sh"

octopus::log::install_errexit

source "${ROOT_DIR}/hack/lib/version.sh"

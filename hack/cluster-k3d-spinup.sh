#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# The root of the octopus directory
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"
source "${ROOT_DIR}/hack/lib/cluster-k3d.sh"

octopus::cluster_k3d::spinup

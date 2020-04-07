#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# The root of the octopus directory
ROOT_DIR="${CURR_DIR}"
source "${ROOT_DIR}/hack/lib/init.sh"
source "${CURR_DIR}/hack/lib/constant.sh"

function entry() {
  local adaptor="${1}"
  if [[ -f "${ROOT_DIR}/adaptors/${adaptor}/Makefile" ]]; then
    make -se -f "${ROOT_DIR}/adaptors/${adaptor}/Makefile" adaptor "$@"
  else
    octopus::log::fatal "could not find '${adaptor}' adaptor !!!"
  fi
}

entry "$@"

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
  local stage="${1:-build}"
  shift $(($# > 0 ? 1 : 0))

  "${ROOT_DIR}/hack/make-rules/octopus.sh" "${stage}" "$@"
  for d in $(octopus::util::find_subdirs "${ROOT_DIR}/adaptors"); do
    "${ROOT_DIR}/hack/make-rules/adaptor.sh" "${d}" "${stage}" "$@"
  done
}

entry "$@"

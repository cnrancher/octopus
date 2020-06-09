#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# The root of the octopus directory
CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"
source "${ROOT_DIR}/hack/lib/cfssl.sh"

if ! octopus::cfssl::validate; then
  octopus::log::error "cannot execute cfssl as it hasn't installed"
  return
fi

# generate client cert key and csr
octopus::log::info "generating client cert key and csr"
if cfssl genkey -config "${ROOT_DIR}/hack/lib/cfssl.json" -profile client "${CURR_DIR}/client_csr.json" | cfssljson -bare "${CURR_DIR}/client"; then
  octopus::log::info "generated client cert key and csr"
else
  octopus::log::fatal "cannot generate client cert key and csr"
fi

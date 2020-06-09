#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# The root of the octopus directory
CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"
source "${ROOT_DIR}/hack/lib/cfssl.sh"

if ! octopus::cfssl::validate; then
  octopus::log::error "cannot execute cfssl as it hasn't installed"
  return
fi

octopus::log::info "
Certificate structure:

CA
 |
 | - - > Intermediate CA
            |
            | - - > Client
"

# generate CA
octopus::log::info "generating CA cert bundle"
if cfssl gencert -initca "${CURR_DIR}/ca_csr.json" | cfssljson -bare "${CURR_DIR}/ca"; then
  octopus::log::info "generated CA cert bundle"
else
  octopus::log::fatal "cannot generate CA cert bundle"
fi

# generate intermediate cert
octopus::log::info "generating Intermediate cert bundle"
if cfssl genkey -initca "${CURR_DIR}/intermediate_csr.json" | cfssljson -bare "${CURR_DIR}/intermediate"; then
  if cfssl sign -config "${ROOT_DIR}/hack/lib/cfssl.json" -profile intermediate -ca "${CURR_DIR}/ca.pem" -ca-key "${CURR_DIR}/ca-key.pem" "${CURR_DIR}/intermediate.csr" | cfssljson -bare "${CURR_DIR}/intermediate"; then
    octopus::log::info "generated Intermediate cert bundle"
  else
    octopus::log::fatal "cannot generate Intermediate cert"
  fi
else
  octopus::log::fatal "cannot generate Intermediate key/csr"
fi

# generate client cert
octopus::log::info "generating Client cert bundle"
if cfssl gencert -config "${ROOT_DIR}/hack/lib/cfssl.json" -profile client -ca "${CURR_DIR}/intermediate.pem" -ca-key "${CURR_DIR}/intermediate-key.pem" "${CURR_DIR}/client_csr.json" | cfssljson -bare "${CURR_DIR}/client"; then
  octopus::log::info "generated Client cert bundle"
else
  octopus::log::fatal "cannot generate Client cert bundle"
fi

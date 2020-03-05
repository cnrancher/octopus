#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Dapper variables helpers. These functions need the
# following variables:
#
#    OS_TYPE          -  The type for the localhost OS, default is automatically discovered.
#    OS_ARCH          -  The arch for the localhost OS, default is automatically discovered.
#    DAPPER_VERSION   -  The dapper version for running, default is v0.4.2.

OS_TYPE=${OS_TYPE:-"$(uname -s)"}
OS_ARCH=${OS_ARCH:-"$(uname -m)"}
DAPPER_VERSION=${DAPPER_VERSION:-"v0.4.2"}

function octopus::dapper::install() {
  curl -sSfL "https://releases.rancher.com/dapper/${DAPPER_VERSION}/dapper-${OS_TYPE}-${OS_ARCH}" >/tmp/dapper
  chmod +x /tmp/dapper && mv /tmp/dapper /usr/local/bin/dapper
}

function octopus::dapper::validate() {
  if [[ -n "$(command -v dapper)" ]]; then
    return 0
  fi

  octopus::log::info "installing dapper (version: ${DAPPER_VERSION}, os: ${OS_TYPE}-${OS_ARCH})"
  if octopus::dapper::install; then
    octopus::log::info "dapper: $(dapper -v)"
    return 0
  fi
  octopus::log::error "no dapper available"
  return 1
}

function octopus::dapper::run() {
  if ! octopus::dapper::validate; then
    octopus::log::fatal "dapper hasn't been installed"
  fi
  dapper "$@"
}

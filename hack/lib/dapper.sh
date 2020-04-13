#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Dapper variables helpers. These functions need the
# following variables:
#
#    DAPPER_VERSION   -  The dapper version for running, default is v0.4.2.

function octopus::dapper::install() {
  local version=${DAPPER_VERSION:-"v0.4.2"}
  curl -fL "https://github.com/rancher/dapper/releases/download/${version}/dapper-$(uname -s)-$(uname -m)" -o /tmp/dapper
  chmod +x /tmp/dapper && mv /tmp/dapper /usr/local/bin/dapper
}

function octopus::dapper::validate() {
  if [[ -n "$(command -v dapper)" ]]; then
    return 0
  fi

  octopus::log::info "installing dapper"
  if octopus::dapper::install; then
    octopus::log::info "dapper: $(dapper -v)"
    return 0
  fi
  octopus::log::error "no dapper available"
  return 1
}

function octopus::dapper::run() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  if ! octopus::dapper::validate; then
    octopus::log::fatal "dapper hasn't been installed"
  fi

  dapper "$@"
}

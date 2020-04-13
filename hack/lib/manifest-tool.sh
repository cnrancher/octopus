#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Manifest tool variables helpers. These functions need the
# following variables:
#    MANIFEST_TOOL_VERSION  -  The manifest tool version for running, default is v0.7.0.
#    DOCKER_USERNAME        -  The username of Docker.
#    DOCKER_PASSWORD        -  The password of Docker.

DOCKER_USERNAME=${DOCKER_USERNAME:-}
DOCKER_PASSWORD=${DOCKER_PASSWORD:-}

function octopus::manifest_tool::install() {
  local version=${MANIFEST_TOOL_VERSION:-"v1.0.1"}
  curl -fL "https://github.com/estesp/manifest-tool/releases/download/${version}/manifest-tool-$(octopus::util::get_os)-$(octopus::util::get_arch ---full-name)" -o /tmp/manifest-tool
  chmod +x /tmp/manifest-tool && mv /tmp/manifest-tool /usr/local/bin/manifest-tool
}

function octopus::manifest_tool::validate() {
  if [[ -n "$(command -v manifest-tool)" ]]; then
    return 0
  fi

  octopus::log::info "installing manifest-tool"
  if octopus::manifest_tool::install; then
    octopus::log::info "$(manifest-tool --version 2>&1)"
    return 0
  fi
  octopus::log::error "no manifest-tool available"
  return 1
}

function octopus::manifest_tool::push() {
  if ! octopus::manifest_tool::validate; then
    octopus::log::error "cannot execute manifest-tool as it hasn't installed"
    return
  fi

  if [[ $(octopus::util::get_os) == "darwin" ]]; then
    if [[ -z ${DOCKER_USERNAME} ]] && [[ -z ${DOCKER_PASSWORD} ]]; then
      # NB(thxCode): since 17.03, Docker for Mac stores credentials in the OSX/macOS keychain and not in config.json, which means the above variables need to specify if using on Mac.
      octopus::log::fatal "must set 'DOCKER_USERNAME' & 'DOCKER_PASSWORD' environment variables in Darwin platform"
    fi
  fi

  octopus::log::info "manifest-tool push $*"
  if [[ -n ${DOCKER_USERNAME} ]] && [[ -n ${DOCKER_PASSWORD} ]]; then
    manifest-tool --username="${DOCKER_USERNAME}" --password="${DOCKER_PASSWORD}" push "$@"
  else
    manifest-tool push "$@"
  fi
}

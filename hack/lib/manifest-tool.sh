#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Manifest tool variables helpers. These functions need the
# following variables:
#    OS_TYPE                -  The type for the localhost OS, default is automatically discovered.
#    OS_ARCH                -  The arch for the localhost OS, default is automatically discovered.
#    MANIFEST_TOOL_VERSION  -  The manifest tool version for running, default is v0.7.0.
#    DOCKER_USERNAME        -  The username of Docker.
#    DOCKER_PASSWORD        -  The password of Docker.

DOCKER_USERNAME=${DOCKER_USERNAME:-}
DOCKER_PASSWORD=${DOCKER_PASSWORD:-}

function octopus::manifest_tool::install() {
  local version=${MANIFEST_TOOL_VERSION:-"v1.0.1"}
  local os_type=${OS_TYPE:-"$(octopus::util::get_os)"}
  local os_arch=${OS_ARCH:-"$(octopus::util::get_arch)"}
  if [[ "${os_arch}" == "arm" ]]; then
    os_arch="armv7"
  fi
  curl -SfL "https://github.com/estesp/manifest-tool/releases/download/${version}/manifest-tool-${os_type}-${os_arch}" >/tmp/manifest-tool
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

function octopus::manifest_tool::run() {
  if ! octopus::manifest_tool::validate; then
    octopus::log::fatal "manifest-tool hasn't been installed"
  fi

  if [[ ${OS_TYPE} == "Darwin" ]]; then
    if [[ -z ${DOCKER_USERNAME} ]] && [[ -z ${DOCKER_PASSWORD} ]]; then
      # NB(thxCode): since 17.03, Docker for Mac stores credentials in the OSX/macOS keychain and not in config.json, which means the above variables need to specify if using on Mac.
      octopus::log::fatal "must set 'DOCKER_USERNAME' & 'DOCKER_PASSWORD' environment variables"
    fi
  fi

  if [[ -n ${DOCKER_USERNAME} ]] && [[ -n ${DOCKER_PASSWORD} ]]; then
    manifest-tool --username="${DOCKER_USERNAME}" --password="${DOCKER_PASSWORD}" "$@"
  else
    manifest-tool "$@"
  fi
}

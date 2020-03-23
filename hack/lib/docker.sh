#!/usr/bin/env bash

function octopus::docker::install() {
  curl -SfL "https://get.docker.com" | sh
}

function octopus::docker::validate() {
  if [[ -n "$(command -v docker)" ]]; then
    return 0
  fi

  octopus::log::info "installing docker"
  if octopus::docker::install; then
    octopus::log::info "docker: $(docker version --format '{{.Server.Version}}' 2>&1)"
    return 0
  fi
  octopus::log::error "no docker available"
  return 1
}

function octopus::docker::build() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  # NB(thxCode): use Docker buildkit to cross build images, ref to:
  # - https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope#buildkit
  DOCKER_BUILDKIT=1 docker build "$@"
}

function octopus::docker::manifest_create() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  # NB(thxCode): use Docker manifest needs to enable client experimental feature, ref to:
  # - https://docs.docker.com/engine/reference/commandline/manifest_create/
  # - https://docs.docker.com/engine/reference/commandline/cli/#experimental-features#environment-variables
  DOCKER_CLI_EXPERIMENTAL=enabled docker manifest create --amend "$@"
}

function octopus::docker::manifest_push() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  # NB(thxCode): use Docker manifest needs to enable client experimental feature, ref to:
  # - https://docs.docker.com/engine/reference/commandline/manifest_push/
  # - https://docs.docker.com/engine/reference/commandline/cli/#experimental-features#environment-variables
  DOCKER_CLI_EXPERIMENTAL=enabled docker manifest push --purge "$@"
}

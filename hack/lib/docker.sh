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

function octopus::docker::login() {
  if [[ -n ${DOCKER_USERNAME} ]] && [[ -n ${DOCKER_PASSWORD} ]]; then
    if ! docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}" >/dev/null 2>&1; then
      return 1
    fi
  fi
  return 0
}

function octopus::docker::build() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  # NB(thxCode): use Docker buildkit to cross build images, ref to:
  # - https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope#buildkit
  DOCKER_BUILDKIT=1 docker build "$@"
}

function octopus::docker::manifest() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  if ! octopus::docker::login; then
    octopus::log::fatal "failed to login docker"
  fi

  # NB(thxCode): use Docker manifest needs to enable client experimental feature, ref to:
  # - https://docs.docker.com/engine/reference/commandline/manifest_create/
  # - https://docs.docker.com/engine/reference/commandline/cli/#experimental-features#environment-variables
  octopus::log::info "docker manifest create --amend $*"
  DOCKER_CLI_EXPERIMENTAL=enabled docker manifest create --amend "$@"

  # NB(thxCode): use Docker manifest needs to enable client experimental feature, ref to:
  # - https://docs.docker.com/engine/reference/commandline/manifest_push/
  # - https://docs.docker.com/engine/reference/commandline/cli/#experimental-features#environment-variables
  octopus::log::info "docker manifest push --purge ${1}"
  DOCKER_CLI_EXPERIMENTAL=enabled docker manifest push --purge "${1}"
}

function octopus::docker::push() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  if ! octopus::docker::login; then
    octopus::log::fatal "failed to login docker"
  fi

  for image in "$@"; do
    octopus::log::info "docker push ${image}"
    docker push "${image}"
  done
}

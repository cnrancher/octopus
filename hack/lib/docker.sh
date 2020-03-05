#!/usr/bin/env bash

function octopus::docker::install() {
  curl -sSfL "https://get.docker.com" | sh
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

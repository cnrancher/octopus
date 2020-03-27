#!/usr/bin/env bash

function octopus::lint::install() {
  curl -fL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.23.8
}

function octopus::lint::validate() {
  if [[ -n "$(command -v golangci-lint)" ]]; then
    return 0
  fi

  octopus::log::info "installing golangci-lint"
  if octopus::lint::install; then
    octopus::log::info "$(golangci-lint --version)"
    return 0
  fi
  octopus::log::warn "no golangci-lint available, using go fmt/vet instead"
  return 1
}

function octopus::lint::generate() {
  if octopus::lint::validate; then
    for path in "$@"; do
      golangci-lint run "${path}"
    done
  else
    octopus::log::warn "no golangci-lint available, using go fmt/vet instead"
    for path in "$@"; do
      go fmt "${path}"
      go vet "${path}"
    done
  fi
}

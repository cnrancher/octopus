#!/usr/bin/env bash

function octopus::controller_gen::install() {
  tmp_dir=$(mktemp -d)
  pushd "${tmp_dir}" >/dev/null || exit 1
  go mod init tmp
  go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
  rm -rf "${tmp_dir}"
  popd >/dev/null || return
}

function octopus::controller_gen::validate() {
  if [[ -n "$(command -v controller-gen)" ]]; then
    return 0
  fi

  octopus::log::info "installing controller-gen"
  if octopus::controller_gen::install; then
    octopus::log::info "controller-gen: $(controller-gen --version)"
    return 0
  fi
  octopus::log::error "no controller-gen available"
  return 1
}

function octopus::controller_gen::generate() {
  if ! octopus::controller_gen::validate; then
    octopus::log::warn "cannot execute controller-gen as it hasn't installed"
    return
  fi
  controller-gen "$@"
}

#!/usr/bin/env bash

function octopus::cfssl::install() {
  tmp_dir=$(mktemp -d)
  pushd "${tmp_dir}" >/dev/null || exit 1
  go mod init tmp
  go get github.com/cloudflare/cfssl/cmd/...
  rm -rf "${tmp_dir}"
  popd >/dev/null || return
}

function octopus::cfssl::validate() {
  if [[ -n "$(command -v cfssl)" ]] || [[ -n "$(command -v cfssljson)" ]]; then
    return 0
  fi

  octopus::log::info "installing cfssl"
  if octopus::cfssl::install; then
    octopus::log::info "$(cfssl version 2>&1)"
    return 0
  fi
  octopus::log::error "no cfssl available"
  return 1
}
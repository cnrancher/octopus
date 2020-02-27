#!/usr/bin/env bash

function octopus::controller_gen::install() {
  tmp_dir=$(mktemp -d)
  pushd "${tmp_dir}" || exit 1
  go mod init tmp
  go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
  rm -rf "${tmp_dir}"
  popd || return
}

function octopus::controller_gen::validate_cli() {
  bin_path="$(command -v controller-gen)"
  if [[ -z "${bin_path}" ]]; then
    octopus::controller_gen::install
    bin_path="${GOPATH}/bin/controller-gen"
  fi
  echo -n "${bin_path}"
}

function octopus::controller_gen::generate() {
  controller_gen_path=$(octopus::controller_gen::validate_cli)
  ${controller_gen_path} "$@"
}

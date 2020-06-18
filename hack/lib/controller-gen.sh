#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Controller-gen variables helpers. These functions need the
# following variables:
#
#    CONTROLLER_GEN_VERSION  -  The go controller-gen version, default is v0.3.0.

function octopus::controller_gen::install() {
  local version=${CONTROLLER_GEN_VERSION:-"v0.3.0"}
  tmp_dir=$(mktemp -d)
  pushd "${tmp_dir}" >/dev/null || exit 1
  go mod init tmp
  GO111MODULE=on go get "sigs.k8s.io/controller-tools/cmd/controller-gen@${version}"
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
    octopus::log::error "cannot execute controller-gen as it hasn't installed"
    return
  fi
  controller-gen "$@"

  # NB(thxCode) remove the `controller-gen.kubebuilder.io/version` annotation from generated YAML files,
  # which is good for updating controller-gen.
  local out_dir=""
  for arg in "$@"; do
    if [[ "${arg}" =~ ^output:.*:dir= ]]; then
      out_dir=${arg#*=}
      break
    fi
  done
  if [[ -d "${out_dir}" ]]; then
    while read -r target_file; do
      if [[ ! -f ${target_file} ]]; then
        continue
      fi
      if ! sed -i 's/controller-gen\.kubebuilder\.io\/version:.*/{}/g' "${target_file}" >/dev/null 2>&1; then
        # back off none GNU sed
        sed -i '' 's/controller-gen\.kubebuilder\.io\/version:.*/{}/g' "${target_file}"
      fi
    done <<<"$(grep -rl "controller-gen.kubebuilder.io/version:" "${out_dir}")"
  fi
}

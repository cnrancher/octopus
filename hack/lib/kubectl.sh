#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Kubectl variables helpers. These functions need the
# following variables:
#
#    K8S_VERSION     -  The Kubernetes version for the cluster, default is v1.17.2.

function octopus::kubectl::install() {
  local version=${K8S_VERSION:-"v1.17.2"}
  curl -fL "https://storage.googleapis.com/kubernetes-release/release/${version}/bin/$(octopus::util::get_os)/$(octopus::util::get_arch)/kubectl" -o /tmp/kubectl
  chmod +x /tmp/kubectl && sudo mv /tmp/kubectl /usr/local/bin/kubectl
}

function octopus::kubectl::validate() {
  if [[ -n "$(command -v kubectl)" ]]; then
    return 0
  fi

  octopus::log::info "installing kubectl"
  if octopus::kubectl::install; then
    octopus::log::info "kubectl: $(kubectl version --short --client)"
    return 0
  fi
  octopus::log::error "no kubectl available"
  return 1
}

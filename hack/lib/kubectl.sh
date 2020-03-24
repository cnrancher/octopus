#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Kubectl variables helpers. These functions need the
# following variables:
#
#    OS_TYPE         -  The type for the localhost OS, default is automatically discovered.
#    OS_ARCH         -  The arch for the localhost OS, default is automatically discovered.
#    K8S_VERSION     -  The Kubernetes version for the cluster, default is v1.17.2.

function octopus::kubectl::install() {
  local version=${K8S_VERSION:-"v1.17.2"}
  local os_type=${OS_TYPE:-"$(octopus::util::get_os)"}
  local os_arch=${OS_ARCH:-"$(octopus::util::get_arch)"}
  curl -fL "https://storage.googleapis.com/kubernetes-release/release/${version}/bin/${os_type}/${os_arch}/kubectl" >/tmp/kubectl
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

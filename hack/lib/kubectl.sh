#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Kubectl variables helpers. These functions need the
# following variables:
#
#    OS_TYPE         -  The type for the localhost OS, default is automatically discovered.
#    OS_ARCH         -  The arch for the localhost OS, default is automatically discovered.
#    K8S_VERSION     -  The Kubernetes version for the cluster, default is v1.17.2.

OS_TYPE=${OS_TYPE:-"$(uname -s)"}
OS_ARCH=${OS_ARCH:-"$(uname -m)"}
K8S_VERSION=${K8S_VERSION:-"v1.17.2"}

function octopus::kubectl::install() {
  local os_type
  os_type=$(echo -n "${OS_TYPE}" | tr '[:upper:]' '[:lower:]')
  local os_arch=${OS_ARCH:-"amd64"}
  if [[ "${os_arch}" == "x86_64" ]]; then
    os_arch="amd64"
  fi
  curl -SfL "https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/${os_type}/${os_arch}/kubectl" >/tmp/kubectl
  chmod +x /tmp/kubectl && sudo mv /tmp/kubectl /usr/local/bin/kubectl
}

function octopus::kubectl::validate() {
  if [[ -n "$(command -v kubectl)" ]]; then
    return 0
  fi

  octopus::log::info "installing kubectl (version: ${K8S_VERSION}, os: ${OS_TYPE}-${OS_ARCH})"
  if octopus::kubectl::install; then
    octopus::log::info "kubectl: $(kubectl version --short --client)"
    return 0
  fi
  octopus::log::error "no kubectl available"
  return 1
}

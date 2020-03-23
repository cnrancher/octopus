#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Kind cluster variables helpers. These functions need the
# following variables:
#
#    OS_TYPE         -  The type for the localhost OS, default is automatically discovered.
#    OS_ARCH         -  The arch for the localhost OS, default is automatically discovered.
#    KIND_VERSION    -  The Kind version for running, default is v0.7.0.
#    K8S_VERSION     -  The Kubernetes version for the cluster, default is v1.17.2.
#    CLUSTER_NAME    -  The name for the cluster, default is octopus-test.
#    CLUSTER_CONFIG  -  The bootstrap configuration path for the cluster, if needed.

OS_TYPE=${OS_TYPE:-"$(uname -s)"}
OS_ARCH=${OS_ARCH:-"$(uname -m)"}
KIND_VERSION=${KIND_VERSION:-"v0.7.0"}
K8S_VERSION=${K8S_VERSION:-"v1.17.2"}
CLUSTER_NAME=${CLUSTER_NAME:-"edge"}
CLUSTER_CONFIG=${CLUSTER_CONFIG:-}
if [[ -z "${CLUSTER_CONFIG}" ]]; then
  CLUSTER_CONFIG="/tmp/default-cluster-config.yaml"
fi

function octopus::cluster_kind::install() {
  local os_type, os_arch
  os_type=$(echo -n "${OS_TYPE}" | tr '[:upper:]' '[:lower:]')
  os_arch=${OS_ARCH:-"amd64"}
  if [[ "${os_arch}" == "x86_64" ]]; then
    os_arch="amd64"
  fi
  curl -SfL "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-${os_type}-${os_arch}" >/tmp/kind
  chmod +x /tmp/kind && mv /tmp/kind /usr/local/bin/kind
}

function octopus::cluster_kind::validate() {
  if [[ -n "$(command -v kind)" ]]; then
    return 0
  fi

  octopus::log::info "installing kind (version: ${KIND_VERSION}, os: ${OS_TYPE}-${OS_ARCH})"
  if octopus::cluster_kind::install; then
    octopus::log::info "kind: $(kind --version 2>&1 | awk '{print $NF}')"
    return 0
  fi
  octopus::log::error "no kind available"
  return 1
}

function octopus::cluster_kind:configure_default() {
  if [[ ! -f "${CLUSTER_CONFIG}" ]]; then
    octopus::log::info "using default cluster config"
    cat >"${CLUSTER_CONFIG}" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
  - role: worker
  - role: worker
EOF
  fi
}

function octopus::cluster_kind::startup() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  if ! octopus::kubectl::validate; then
    octopus::log::fatal "kubectl hasn't been installed"
  fi
  if ! octopus::cluster_kind::validate; then
    octopus::log::fatal "kind hasn't been installed"
  fi

  octopus::log::info "creating ${CLUSTER_NAME} cluster with ${K8S_VERSION}"
  octopus::cluster_kind:configure_default
  kind create cluster --name "${CLUSTER_NAME}" --config "${CLUSTER_CONFIG}" --image="kindest/node:${K8S_VERSION}" --wait 5m

  local kubconfig="/tmp/kubeconfig-${CLUSTER_NAME}.yaml"
  octopus::log::info "exporting ${CLUSTER_NAME} cluster's kubeconfig to ${kubconfig}"
  rm -f "${kubconfig}"
  kind export kubeconfig --name "${CLUSTER_NAME}" --kubeconfig "${kubconfig}"
}

function octopus::cluster_kind::cleanup() {
  octopus::log::warn "removing ${CLUSTER_NAME} cluster"
  kind delete cluster --name "${CLUSTER_NAME}"
}

function octopus::cluster_kind::spinup() {
  trap 'octopus::cluster_kind::cleanup' EXIT
  octopus::cluster_kind::startup

  octopus::log::warn "please input CTRL+C to stop the local cluster"
  read -r /dev/null
}

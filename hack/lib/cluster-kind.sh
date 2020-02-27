#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Kind cluster variables helpers. These functions need the
# following variables:
#
#    OS_ARCH - The arch for the localhost OS, default is amd64.
#    OS_TYPE - The type for the localhost OS, default is automatically discovered.
#    CLUSTER_NAME - The name for the cluster, default is octopus-test.
#    CLUSTER_VERSION - The Kubernetes version for the cluster, default is v1.16.3.
#    CLUSTER_CONFIG - The bootstrap configuration path for the cluster, if needed.

OS_ARCH=${OS_ARCH:-"amd64"}
OS_TYPE=${OS_TYPE:-}
if [[ "${OS_TYPE}x" == "x" ]]; then
  OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
fi
CLUSTER_NAME=${CLUSTER_NAME:-"edge"}
CLUSTER_VERSION=${CLUSTER_VERSION:-"v1.17.2"}
CLUSTER_CONFIG=${CLUSTER_CONFIG:-}
if [[ "${CLUSTER_CONFIG}x" == "x" ]]; then
  CLUSTER_CONFIG="/tmp/default-cluster-config.yaml"
fi
KIND_VERSION=${KIND_VERSION:-"v0.7.0"}

function octopus::cluster_kind::install_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    octopus::log::warn "installing docker"
    curl -L "https://get.docker.com" | sh
  fi

  docker_version="$(docker version --format '{{.Server.Version}}' 2>&1)"
  octopus::log::info "docker: ${docker_version}"
}

function octopus::cluster_kind::install_kubectl() {
  if ! command -v kubectl >/dev/null 2>&1; then
    octopus::log::warn "installing kubectl"
    local k8s_version
    k8s_version=$(curl -s "https://storage.googleapis.com/kubernetes-release/release/stable.txt")
    curl -L "https://storage.googleapis.com/kubernetes-release/release/${k8s_version}/bin/${OS_TYPE}/${OS_ARCH}/kubectl" >/tmp/kubectl
    chmod +x /tmp/kubectl && sudo mv /tmp/kubectl /usr/local/bin/kubectl
  fi
  octopus::log::info "kubectl: $(kubectl version --short --client)"
}

function octopus::cluster_kind::install() {
  if ! command -v kind >/dev/null 2>&1; then
    octopus::log::warn "installing kind"
    curl -L "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-${OS_TYPE}-${OS_ARCH}" >/tmp/kind
    chmod +x /tmp/kind && sudo mv /tmp/kind /usr/local/bin/kind
  fi
  octopus::log::info "kind: $(kind --version 2>&1 | awk '{print $NF}')"
}

function octopus::cluster_kind:configure_default() {
  if [[ ! -f "${CLUSTER_CONFIG}" ]]; then
    octopus::log::info "using default cluster config"
    cat >"${CLUSTER_CONFIG}" <<EOF
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
nodes:
  - role: control-plane
  - role: worker
  - role: worker
  - role: worker
EOF
  fi
}

function octopus::cluster_kind::startup() {
  octopus::cluster_kind::install_docker
  octopus::cluster_kind::install_kubectl
  octopus::cluster_kind::install

  octopus::log::info "creating ${CLUSTER_NAME} cluster with ${CLUSTER_VERSION}"
  octopus::cluster_kind:configure_default
  kind create cluster --name "${CLUSTER_NAME}" --config "${CLUSTER_CONFIG}" --image="kindest/node:${CLUSTER_VERSION}" --wait 5m

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

#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Kind cluster variables helpers. These functions need the
# following variables:
#
#    OS_TYPE         -  The type for the localhost OS, default is automatically discovered.
#    OS_ARCH         -  The arch for the localhost OS, default is automatically discovered.
#    KIND_VERSION    -  The Kind version for running, default is v0.7.0.
#    K8S_VERSION     -  The Kubernetes version for the cluster, default is v1.17.2.
#    CLUSTER_NAME    -  The name for the cluster, default is edge.
#    CONTROL_PLANES  -  The number of the control-plane, default is 1.
#    WORKERS         -  The number of the workers, default is 3.

K8S_VERSION=${K8S_VERSION:-"v1.17.2"}
CLUSTER_NAME=${CLUSTER_NAME:-"edge"}

function octopus::cluster_kind::install() {
  local version=${KIND_VERSION:-"v1.17.2"}
  local os_type=${OS_TYPE:-"$(octopus::util::get_os)"}
  local os_arch=${OS_ARCH:-"$(octopus::util::get_arch)"}
  curl -fL "https://github.com/kubernetes-sigs/kind/releases/download/${version}/kind-${os_type}-${os_arch}" >/tmp/kind
  chmod +x /tmp/kind && mv /tmp/kind /usr/local/bin/kind
}

function octopus::cluster_kind::validate() {
  if [[ -n "$(command -v kind)" ]]; then
    return 0
  fi

  octopus::log::info "installing kind"
  if octopus::cluster_kind::install; then
    octopus::log::info "$(kind --version 2>&1)"
    return 0
  fi
  octopus::log::error "no kind available"
  return 1
}

function octopus::cluster_kind:setup_configuration() {
  local config="/tmp/kind-${CLUSTER_NAME}-cluster-config.yaml"
  cat >"${config}" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
EOF

  local control_planes=${CONTROL_PLANES:-1}
  if [[ ${control_planes} -lt 1 ]]; then
    control_planes=1
  fi
  for ((i = 0; i < control_planes; i++)); do
    # shellcheck disable=SC2086
    cat >>${config} <<EOF
  - role: control-plane
EOF
  done

  local workers=${WORKERS:-3}
  if [[ ${workers} -lt 1 ]]; then
    workers=1
  fi
  for ((i = 0; i < workers; i++)); do
    cat >>"${config}" <<EOF
  - role: worker
EOF
  done

  echo -n "${config}"
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
  # setup cluster
  local config
  config=$(octopus::cluster_kind:setup_configuration)
  local kind_image="kindest/node:${K8S_VERSION}"
  kind create cluster --name "${CLUSTER_NAME}" --config "${config}" --image "${kind_image}" --wait 5m

  octopus::log::info "${CLUSTER_NAME} cluster's kubeconfig has wrote in the ~/.kube/config"
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

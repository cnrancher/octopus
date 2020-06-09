#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Kind cluster variables helpers. These functions need the
# following variables:
#
#    KIND_VERSION    -  The Kind version for running, default is v0.8.1.
#    K8S_VERSION     -  The Kubernetes version for the cluster, default is v1.18.2.
#    CLUSTER_NAME    -  The name for the cluster, default is edge.
#    CONTROL_PLANES  -  The number of the control-plane, default is 1.
#    WORKERS         -  The number of the workers, default is 3.

K8S_VERSION=${K8S_VERSION:-"v1.18.2"}
CLUSTER_NAME=${CLUSTER_NAME:-"edge"}

function octopus::cluster_kind::install() {
  local version=${KIND_VERSION:-"v0.8.1"}
  curl -fL "https://github.com/kubernetes-sigs/kind/releases/download/${version}/kind-$(octopus::util::get_os)-$(octopus::util::get_arch)" -o /tmp/kind
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

function octopus::cluster_kind::setup_configuration() {
  local config="/tmp/kind-${CLUSTER_NAME}-cluster-config.yaml"
  cat >"${config}" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: "0.0.0.0"
nodes:
EOF

  local control_planes=${CONTROL_PLANES:-1}
  if [[ ${control_planes} -lt 1 ]]; then
    control_planes=1
  fi
  for ((i = 0; i < control_planes; i++)); do
    if [[ ${i} -eq 0 ]]; then
      local random_port_start
      random_port_start=$(octopus::util::get_random_port_start 2)
      local ingress_http_port=$((random_port_start + 0))
      octopus::log::info "INGRESS_HTTP_PORT is ${ingress_http_port}"
      export INGRESS_HTTP_PORT=${ingress_http_port}
      local ingress_https_port=$((random_port_start + 1))
      octopus::log::info "INGRESS_HTTPS_PORT is ${ingress_https_port}"
      export INGRESS_HTTPS_PORT=${ingress_https_port}

      # shellcheck disable=SC2086
      cat >>${config} <<EOF
  - role: control-plane
    kubeadmConfigPatches:
    - |
      kind: InitConfiguration
      nodeRegistration:
        kubeletExtraArgs:
          node-labels: "ingress-ready=true"
    extraPortMappings:
    - containerPort: 80
      hostPort: ${ingress_http_port}
      protocol: TCP
    - containerPort: 443
      hostPort: ${ingress_https_port}
      protocol: TCP
EOF
    else
      # shellcheck disable=SC2086
      cat >>${config} <<EOF
  - role: control-plane
EOF
    fi
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
  octopus::cluster_kind::setup_configuration
  local kind_image="kindest/node:${K8S_VERSION}"
  kind create cluster --name "${CLUSTER_NAME}" --config "/tmp/kind-${CLUSTER_NAME}-cluster-config.yaml" --image "${kind_image}" --wait 5m

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

function octopus::cluster_kind::add_worker() {
  if ! octopus::docker::validate; then
    octopus::log::fatal "docker hasn't been installed"
  fi
  if ! octopus::kubectl::validate; then
    octopus::log::fatal "kubectl hasn't been installed"
  fi
  if ! octopus::cluster_kind::validate; then
    octopus::log::fatal "kind hasn't been installed"
  fi

  octopus::log::error "there is not add-node operation in kind"
}

#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# The root of the octopus directory
ROOT_DIR="$(cd "${CURR_DIR}/../.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"

function generate() {
  local adaptor="${1}"

  octopus::log::info "generating adaptor $adaptor..."

  octopus::log::info "generating objects"
  rm -f "${CURR_DIR}/api/*/zz_generated*"
  octopus::controller_gen::generate \
    object:headerFile="${ROOT_DIR}/hack/boilerplate.go.txt" \
    paths="${CURR_DIR}/api/..."

  octopus::log::info "generating manifests"
  # generate crd
  octopus::controller_gen::generate \
    crd:crdVersions=v1 \
    paths="${CURR_DIR}/api/..." \
    output:crd:dir="${CURR_DIR}/deploy/manifests/crd/base"
  # generate rbac role
  octopus::controller_gen::generate \
    rbac:roleName=manager-role \
    paths="${CURR_DIR}/pkg/..." \
    output:rbac:dir="${CURR_DIR}/deploy/manifests/rbac"

  octopus::log::info "merging manifests"
  if ! octopus::kubectl::validate; then
    octopus::log::fatal "kubectl hasn't been installed"
  fi
  kubectl kustomize "${CURR_DIR}/deploy/manifests/overlays/default" \
    >"${CURR_DIR}/deploy/e2e/all_in_one.yaml"

  octopus::log::info "...done"
}

function mod() {
  local adaptor="${1}"

  # the adaptor is sharing the vendor with root
  pushd "${ROOT_DIR}" >/dev/null || exist 1
  octopus::log::info "downloading dependencies for adaptor $adaptor..."

  octopus::log::info "tidying"
  go mod tidy

  octopus::log::info "vending"
  go mod vendor

  octopus::log::info "...done"
  popd >/dev/null || return
}

function lint() {
  local adaptor="${1}"

  octopus::log::info "linting adaptor $adaptor..."

  octopus::lint::generate "${CURR_DIR}/..."

  octopus::log::info "...done"
}

function build() {
  if [[ "${CURR_DIR}" =~ "$(go env GOPATH)".* ]]; then
    export GO111MODULE=off
  fi

  local adaptor="${1}"

  octopus::log::info "building adaptor $adaptor..."

  local version_flags="
    -X k8s.io/client-go/pkg/version.gitVersion=${OCTOPUS_GIT_VERSION}
    -X k8s.io/client-go/pkg/version.gitCommit=${OCTOPUS_GIT_COMMIT}
    -X k8s.io/client-go/pkg/version.gitTreeState=${OCTOPUS_GIT_TREE_STATE}
    -X k8s.io/client-go/pkg/version.buildDate=${OCTOPUS_BUILD_DATE}"
  local flags="
    -w -s"
  local ext_flags="
    -extldflags '-static'"
  local os
  os=$(go env GOOS)
  local arch
  arch=$(go env GOARCH)

  mkdir -p "${CURR_DIR}/bin"
  CGO_ENABLED=0 go build \
    -ldflags "${version_flags} ${flags} ${ext_flags}" \
    -o "${CURR_DIR}/bin/${adaptor}_${os}_${arch}" \
    "${CURR_DIR}/cmd/${adaptor}/main.go"
  mkdir -p "${CURR_DIR}/dist"
  cp -f "${CURR_DIR}/bin/${adaptor}_${os}_${arch}" "${CURR_DIR}/dist/${adaptor}"

  octopus::log::info "...done"
}

function test() {
  if [[ "${CURR_DIR}" =~ "$(go env GOPATH)".* ]]; then
    export GO111MODULE=off
  fi

  local adaptor="${1}"

  octopus::log::info "running unit tests for adaptor $adaptor..."

  local unit_test_targets=(
    "${CURR_DIR}/api/..."
    "${CURR_DIR}/cmd/..."
    "${CURR_DIR}/pkg/..."
  )
  local os
  os=$(go env GOOS)
  local arch
  arch=$(go env GOARCH)

  CGO_ENABLED=1 go test \
    -race \
    -cover -coverprofile "${CURR_DIR}/dist/coverage_${os}_${arch}.out" \
    "${unit_test_targets[@]}"

  octopus::log::info "...done"
}

function verify() {
  if [[ "${CURR_DIR}" =~ "$(go env GOPATH)".* ]]; then
    export GO111MODULE=off
  fi

  local adaptor="${1}"

  octopus::log::info "running integration tests for adaptor $adaptor..."

  octopus::log::info "...done"
}

function containerize() {
  local adaptor="${1}"

  local suffix="-${ARCH:-$(go env GOARCH)}"
  local tag=${TAG:-${OCTOPUS_GIT_VERSION}${suffix}}
  local repo=${REPO:-rancher}
  local image_name=${IMAGE_NAME:-octopus-adaptor-${adaptor}}
  local image=${repo}/${image_name}:${tag}

  octopus::log::info "containerizing adaptor ${adaptor} in ${image}"

  docker build -t "${image}" -f "${CURR_DIR}/Dockerfile" .

  octopus::log::info "...done"
}

function package() {
  local adaptor="${1}"

  octopus::log::info "packaging adaptor $adaptor..."

  octopus::dapper::run -C "${ROOT_DIR}" -f "${CURR_DIR}/Dockerfile.dapper" -m bind

  octopus::log::info "...done"
}

function e2e() {
  if [[ "${CURR_DIR}" =~ "$(go env GOPATH)".* ]]; then
    export GO111MODULE=off
  fi

  local adaptor="${1}"

  octopus::log::info "running E2E tests for adaptor $adaptor..."

  octopus::log::info "...done"
}

function deploy() {
  local adaptor="${1}"

  octopus::log::info "deploying adaptor $adaptor..."

  octopus::log::info "...done"
}

function entry() {
  local adaptor
  adaptor="${1}"

  local stage
  stage="${2-:build}"

  local subcmd
  subcmd="${3-:}"

  case $stage in
  g | gen | generate)
    generate "${adaptor}"
    ;;
  m | mod)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
    fi
    mod "${adaptor}"
    ;;
  l | lint)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
      mod "${adaptor}"
    fi
    lint "${adaptor}"
    ;;
  b | build)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
      mod "${adaptor}"
      lint "${adaptor}"
    fi
    build "${adaptor}"
    ;;
  t | test)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
      mod "${adaptor}"
      lint "${adaptor}"
      build "${adaptor}"
    fi
    test "${adaptor}"
    ;;
  v | verify)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
      mod "${adaptor}"
      lint "${adaptor}"
      build "${adaptor}"
      test "${adaptor}"
    fi
    verify "${adaptor}"
    ;;
  p | pkg | package)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
      mod "${adaptor}"
      lint "${adaptor}"
    fi
    octopus::dapper::run -C "${ROOT_DIR}" -f "adaptors/${adaptor}/Dockerfile.dapper" -m bind
    ;;
  c | containerize)
    if [[ "${subcmd}" != "only" ]]; then
      build "${adaptor}"
      test "${adaptor}"
    fi
    containerize "${adaptor}"
    ;;
  e | e2e)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
      mod "${adaptor}"
      lint "${adaptor}"
      package "${adaptor}"
    fi
    e2e "${adaptor}"
    ;;
  d | deploy)
    if [[ "${subcmd}" != "only" ]]; then
      generate "${adaptor}"
      mod "${adaptor}"
      lint "${adaptor}"
      package "${adaptor}"
      e2e "${adaptor}"
    fi
    deploy "${adaptor}"
    ;;
  *)
    octopus::log::error "unknown action, select from (generate,mod,lint,build,test,verify,package,containerize,e2e,deploy) "
    ;;
  esac
}

entry "$@"

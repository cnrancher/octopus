#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# The root of the octopus directory
ROOT_DIR="${CURR_DIR}"
source "${ROOT_DIR}/hack/lib/init.sh"

function generate() {
  octopus::log::info "generating octopus..."

  octopus::log::info "generating objects"
  rm -f "${CURR_DIR}/api/*/zz_generated*"
  octopus::controller_gen::generate \
    object:headerFile="${ROOT_DIR}/hack/boilerplate.go.txt" \
    paths="${CURR_DIR}/api/..."

  octopus::log::info "generating protos"
  rm -f "${CURR_DIR}/pkg/adaptor/api/*/*.pb.go"
  for d in $(octopus::find::subdirs "${CURR_DIR}/pkg/adaptor/api"); do
    octopus::protoc::generate \
      "${CURR_DIR}/pkg/adaptor/api/${d}"
  done

  octopus::log::info "generating manifests"
  # generate crd
  octopus::controller_gen::generate \
    crd:crdVersions=v1 \
    paths="${CURR_DIR}/api/..." \
    output:crd:dir="${CURR_DIR}/deploy/manifests/crd/base"
  # generate webhook
  octopus::controller_gen::generate \
    webhook \
    paths="${CURR_DIR}/api/..." \
    output:webhook:dir="${CURR_DIR}/deploy/manifests/overlays/webhook"
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
  kubectl kustomize "${CURR_DIR}/deploy/manifests/overlays/webhook" \
    >"${CURR_DIR}/deploy/e2e/all_in_one_with_wehbook.yaml"

  octopus::log::info "...done"
}

function mod() {
  pushd "${ROOT_DIR}" >/dev/null || exist 1
  octopus::log::info "downloading dependencies for octopus..."

  octopus::log::info "tidying"
  go mod tidy

  octopus::log::info "vending"
  go mod vendor

  octopus::log::info "...done"
  popd >/dev/null || return
}

function lint() {
  octopus::log::info "linting octopus..."

  local targets=(
    "${CURR_DIR}/api/..."
    "${CURR_DIR}/cmd/..."
    "${CURR_DIR}/pkg/..."
    "${CURR_DIR}/test/..."
  )
  octopus::lint::generate "${targets[@]}"

  octopus::log::info "...done"
}

function build() {
  if [[ "${CURR_DIR}" =~ "$(go env GOPATH)".* ]]; then
    export GO111MODULE=off
  fi

  octopus::log::info "building octopus..."

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
    -o "${CURR_DIR}/bin/octopus_${os}_${arch}" \
    "${CURR_DIR}/cmd/octopus/main.go"
  mkdir -p "${CURR_DIR}/dist"
  cp -f "${CURR_DIR}/bin/octopus_${os}_${arch}" "${CURR_DIR}/dist/octopus"

  octopus::log::info "...done"
}

function test() {
  if [[ "${CURR_DIR}" =~ "$(go env GOPATH)".* ]]; then
    export GO111MODULE=off
  fi

  octopus::log::info "running unit tests for octopus..."

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

  octopus::log::info "running integration tests for octopus..."

  CGO_ENABLED=1 go test \
    "${CURR_DIR}/test/integration/brain/..."
  #  CGO_ENABLED=1 go test \
  #    "${CURR_DIR}/test/integration/limb/..."

  octopus::log::info "...done"
}

function containerize() {
  local suffix="-${ARCH:-$(go env GOARCH)}"
  local tag=${TAG:-${OCTOPUS_GIT_VERSION}${suffix}}
  local repo=${REPO:-rancher}
  local image_name=${IMAGE_NAME:-octopus}
  local image=${repo}/${image_name}:${tag}

  octopus::log::info "containerizing octopus in ${image}"

  docker build -t "${image}" -f "${CURR_DIR}/Dockerfile" .

  octopus::log::info "...done"
}

function e2e() {
  if [[ "${CURR_DIR}" =~ "$(go env GOPATH)".* ]]; then
    export GO111MODULE=off
  fi

  octopus::log::info "running E2E tests for octopus..."

  octopus::log::info "...done"
}

function deploy() {
  octopus::log::info "deploying octopus..."

  octopus::log::info "...done"
}

function entry() {
  local stage
  stage="${1-:build}"

  local subcmd
  subcmd="${2-:}"

  case $stage in
  g | gen | generate)
    generate
    ;;
  m | mod)
    if [[ "${subcmd}" != "only" ]]; then
      generate
    fi
    mod
    ;;
  l | lint)
    if [[ "${subcmd}" != "only" ]]; then
      generate
      mod
    fi
    lint
    ;;
  b | build)
    if [[ "${subcmd}" != "only" ]]; then
      generate
      mod
      lint
    fi
    build
    ;;
  t | test)
    if [[ "${subcmd}" != "only" ]]; then
      generate
      mod
      lint
      build
    fi
    test
    ;;
  v | verify)
    if [[ "${subcmd}" != "only" ]]; then
      generate
      mod
      lint
      build
      test
    fi
    verify
    ;;
  p | pkg | package)
    if [[ "${subcmd}" != "only" ]]; then
      generate
      mod
      lint
    fi
    octopus::dapper::run -C "${CURR_DIR}" -f "Dockerfile.dapper" -m bind
    ;;
  c | containerize)
    if [[ "${subcmd}" != "only" ]]; then
      build
      test
    fi
    containerize
    ;;
  e | e2e)
    if [[ "${subcmd}" != "only" ]]; then
      generate
      mod
      lint
      package
    fi
    e2e
    ;;
  d | deploy)
    if [[ "${subcmd}" != "only" ]]; then
      generate
      mod
      lint
      package
      e2e
    fi
    deploy
    ;;
  *)
    octopus::log::error "unknown action, select from (generate,mod,lint,build,test,verify,package,containerize,e2e,deploy) "
    ;;
  esac
}

entry "$@"

#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# The root of the octopus directory
ROOT_DIR="$(cd "${CURR_DIR}/../.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"
source "${CURR_DIR}/hack/lib/constant.sh"

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
  local adaptor="${1}"

  octopus::log::info "building adaptor $adaptor..."

  mkdir -p "${CURR_DIR}/bin"

  local version_flags="
    -X k8s.io/client-go/pkg/version.gitVersion=${OCTOPUS_GIT_VERSION}
    -X k8s.io/client-go/pkg/version.gitCommit=${OCTOPUS_GIT_COMMIT}
    -X k8s.io/client-go/pkg/version.gitTreeState=${OCTOPUS_GIT_TREE_STATE}
    -X k8s.io/client-go/pkg/version.buildDate=${OCTOPUS_BUILD_DATE}"
  local flags="
    -w -s"
  local ext_flags="
    -extldflags '-static'"
  local os="${OS:-$(go env GOOS)}"
  local arch="${ARCH:-$(go env GOARCH)}"

  local platforms
  if [[ "${CROSS:-false}" == "true" ]]; then
    octopus::log::info "crossed building"
    platforms=("${SUPPORTED_PLATFORMS[@]}")
  else
    local os="${OS:-$(go env GOOS)}"
    local arch="${ARCH:-$(go env GOARCH)}"
    platforms=("${os}/${arch}")
  fi

  for platform in "${platforms[@]}"; do
    octopus::log::info "building ${platform}"

    local os_arch
    IFS="/" read -r -a os_arch <<<"${platform}"

    local os=${os_arch[0]}
    local arch=${os_arch[1]}
    GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build \
      -ldflags "${version_flags} ${flags} ${ext_flags}" \
      -o "${CURR_DIR}/bin/${adaptor}_${os}_${arch}" \
      "${CURR_DIR}/cmd/${adaptor}/main.go"
  done

  octopus::log::info "...done"
}

function test() {
  local adaptor="${1}"

  octopus::log::info "running unit tests for adaptor $adaptor..."

  local unit_test_targets=(
    "${CURR_DIR}/api/..."
    "${CURR_DIR}/cmd/..."
    "${CURR_DIR}/pkg/..."
  )

  if [[ "${CROSS:-false}" == "true" ]]; then
    octopus::log::warn "crossed test is not supported"
  fi

  local os="${OS:-$(go env GOOS)}"
  local arch="${ARCH:-$(go env GOARCH)}"
  if [[ "${arch}" == "arm" ]]; then
    # NB(thxCode): race detector doesn't support `arm` arch, ref to:
    # - https://golang.org/doc/articles/race_detector.html#Supported_Systems
    GOOS=${os} GOARCH=${arch} CGO_ENABLED=1 go test \
      -cover -coverprofile "${CURR_DIR}/dist/coverage_${adaptor}_${os}_${arch}.out" \
      "${unit_test_targets[@]}"
  else
    GOOS=${os} GOARCH=${arch} CGO_ENABLED=1 go test \
      -race \
      -cover -coverprofile "${CURR_DIR}/dist/coverage_${adaptor}_${os}_${arch}.out" \
      "${unit_test_targets[@]}"
  fi

  octopus::log::info "...done"
}

function verify() {
  local adaptor="${1}"

  octopus::log::info "running integration tests for adaptor $adaptor..."

  octopus::log::info "...done"
}

function containerize() {
  local adaptor="${1}"

  octopus::log::info "containerizing adaptor ${adaptor}..."

  local repo=${REPO:-rancher}
  local image_name=${IMAGE_NAME:-octopus-adaptor-${adaptor}}
  local tag=${TAG:-${OCTOPUS_GIT_VERSION}}

  local platforms
  if [[ "${CROSS:-false}" == "true" ]]; then
    octopus::log::info "crossed containerizing"
    platforms=("${SUPPORTED_PLATFORMS[@]}")
  else
    local os="${OS:-$(go env GOOS)}"
    local arch="${ARCH:-$(go env GOARCH)}"
    platforms=("${os}/${arch}")
  fi

  pushd "${CURR_DIR}"
  for platform in "${platforms[@]}"; do
    octopus::log::info "containerizing ${platform}"
    octopus::docker::build \
      --platform "${platform}" \
      -t "${repo}/${image_name}:${tag}-${platform////-}" .
  done
  popd

  octopus::log::info "...done"
}

function e2e() {
  local adaptor="${1}"

  octopus::log::info "running E2E tests for adaptor $adaptor..."

  octopus::log::info "...done"
}

function deploy() {
  local adaptor="${1}"

  octopus::log::info "deploying adaptor $adaptor..."

  local repo=${REPO:-rancher}
  local image_name=${IMAGE_NAME:-octopus-adaptor-${adaptor}}
  local tag=${TAG:-${OCTOPUS_GIT_VERSION}}
  local images=()
  for platform in "${SUPPORTED_PLATFORMS[@]}"; do
    images+=("${repo}/${image_name}:${tag}-${platform////-}")
  done

  # docker login
  if [[ -n ${DOCKER_USERNAME} ]] && [[ -n ${DOCKER_PASSWORD} ]]; then
    docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"
  fi

  # docker push
  for image in "${images[@]}"; do
    octopus::log::info "deploying image ${image}"
    docker push "${image}"
  done

  # docker manifest
  local targets=(
    "${repo}/${image_name}:${tag}"
    "${repo}/${image_name}:latest"
  )
  for target in "${targets[@]}"; do
    octopus::log::info "deploying manifest image ${target}"
    octopus::docker::manifest_create "${target}" "${images[@]}"
    octopus::docker::manifest_push "${target}"
  done

  octopus::log::info "...done"
}

function entry() {
  local adaptor
  adaptor="${1:-}"
  if [[ -z "${adaptor}" ]]; then
    octopus::log::error "please indicate a specific adaptor !!!"
    exit 0
  elif [[ ! -d "${ROOT_DIR}/adaptors/${adaptor}" ]]; then
    octopus::log::error "does not exist ${adaptor} adaptor !!!"
    exit 0
  fi

  local stage
  stage="${2:-package}"

  local subcmd
  subcmd="${3:-}"

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
    octopus::log::error "unknown action, select from (generate,mod,lint,build,test,verify,containerize,package,e2e,deploy) "
    ;;
  esac
}

entry "$@"

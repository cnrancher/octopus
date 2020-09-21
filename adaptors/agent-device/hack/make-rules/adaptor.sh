#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# The root of the octopus directory
ROOT_DIR="$(cd "${CURR_DIR}/../.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"
source "${CURR_DIR}/hack/lib/constant.sh"

mkdir -p "${CURR_DIR}/bin"
mkdir -p "${CURR_DIR}/dist"

function generate() {
  local adaptor="${1}"

  octopus::log::info "generating adaptor ${adaptor}..."

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
  kustomize build "${CURR_DIR}/deploy/manifests/overlays/default" \
    >"${CURR_DIR}/deploy/e2e/all_in_one.yaml"

  octopus::log::info "...done"
}

function mod() {
  [[ "${2:-}" != "only" ]] && generate "$@"
  local adaptor="${1}"

  # the adaptor is sharing the vendor with root
  pushd "${ROOT_DIR}" >/dev/null || exist 1
  octopus::log::info "downloading dependencies for adaptor ${adaptor}..."

  if [[ "$(go env GO111MODULE)" == "off" ]]; then
    octopus::log::warn "go mod has been disabled by GO111MODULE=off"
  else
    octopus::log::info "tidying"
    go mod tidy
    octopus::log::info "vending"
    go mod vendor
  fi

  octopus::log::info "...done"
  popd >/dev/null || return
}

function lint() {
  [[ "${2:-}" != "only" ]] && mod "$@"
  local adaptor="${1}"

  octopus::log::info "linting adaptor ${adaptor}..."
  octopus::lint::generate "${CURR_DIR}/..."
  octopus::log::info "...done"
}

function build() {
  [[ "${2:-}" != "only" ]] && lint "$@"
  local adaptor="${1}"

  octopus::log::info "building adaptor ${adaptor}(${GIT_VERSION},${GIT_COMMIT},${GIT_TREE_STATE},${BUILD_DATE})..."

  # TODO adjust the ldflags if needed
  local version_flags="
    -X k8s.io/client-go/pkg/version.gitVersion=${GIT_VERSION}
    -X k8s.io/client-go/pkg/version.gitCommit=${GIT_COMMIT}
    -X k8s.io/client-go/pkg/version.gitTreeState=${GIT_TREE_STATE}
    -X k8s.io/client-go/pkg/version.buildDate=${BUILD_DATE}"
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

function package() {
  [[ "${2:-}" != "only" ]] && build "$@"
  local adaptor="${1}"

  octopus::log::info "packaging adaptor ${adaptor}..."

  local repo=${REPO:-rancher}
  local image_name=${IMAGE_NAME:-octopus-adaptor-${adaptor}}
  local tag=${TAG:-${GIT_VERSION}}

  local platforms
  if [[ "${CROSS:-false}" == "true" ]]; then
    octopus::log::info "crossed packaging"
    platforms=("${SUPPORTED_PLATFORMS[@]}")
  else
    local os="${OS:-$(go env GOOS)}"
    local arch="${ARCH:-$(go env GOARCH)}"
    platforms=("${os}/${arch}")
  fi

  pushd "${CURR_DIR}" >/dev/null 2>&1
  for platform in "${platforms[@]}"; do
    if [[ "${platform}" =~ darwin/* ]]; then
      octopus::log::fatal "package into Darwin OS image is unavailable, please use CROSS=true env to containerize multiple arch images or use OS=linux ARCH=amd64 env to containerize linux/amd64 image"
    fi

    local image_tag="${repo}/${image_name}:${tag}-${platform////-}"
    octopus::log::info "packaging ${image_tag}"
    octopus::docker::build \
      --platform "${platform}" \
      -t "${image_tag}" .
  done
  popd >/dev/null 2>&1

  octopus::log::info "...done"
}

function deploy() {
  [[ "${2:-}" != "only" ]] && package "$@"
  local adaptor="${1}"

  octopus::log::info "deploying adaptor ${adaptor}..."

  local repo=${REPO:-rancher}
  local image_name=${IMAGE_NAME:-octopus-adaptor-${adaptor}}
  local tag=${TAG:-${GIT_VERSION}}

  local platforms
  if [[ "${CROSS:-false}" == "true" ]]; then
    octopus::log::info "crossed deploying"
    platforms=("${SUPPORTED_PLATFORMS[@]}")
  else
    local os="${OS:-$(go env GOOS)}"
    local arch="${ARCH:-$(go env GOARCH)}"
    platforms=("${os}/${arch}")
  fi
  local images=()
  for platform in "${platforms[@]}"; do
    if [[ "${platform}" =~ darwin/* ]]; then
      octopus::log::fatal "package into Darwin OS image is unavailable, please use CROSS=true env to containerize multiple arch images or use OS=linux ARCH=amd64 env to containerize linux/amd64 image"
    fi

    images+=("${repo}/${image_name}:${tag}-${platform////-}")
  done

  local only_manifest=${ONLY_MANIFEST:-false}
  local without_manifest=${WITHOUT_MANIFEST:-false}
  local ignore_missing=${IGNORE_MISSING:-false}

  # docker push
  if [[ "${only_manifest}" == "false" ]]; then
    octopus::docker::push "${images[@]}"
  else
    octopus::log::warn "deploying images has been stopped by ONLY_MANIFEST"
    # execute manifest forcibly
    without_manifest="false"
  fi

  # docker manifest
  if [[ "${without_manifest}" == "false" ]]; then
    if [[ "${ignore_missing}" == "false" ]]; then
      octopus::docker::manifest "${repo}/${image_name}:${tag}" "${images[@]}"
    else
      octopus::manifest_tool::push from-args \
        --ignore-missing \
        --target="${repo}/${image_name}:${tag}" \
        --template="${repo}/${image_name}:${tag}-OS-ARCH" \
        --platforms="$(octopus::util::join_array "," "${platforms[@]}")"
    fi

    # generate tested yaml
    local tmpfile
    tmpfile=$(mktemp)
    cp -f "${CURR_DIR}/deploy/e2e/all_in_one.yaml" "${CURR_DIR}/dist/octopus_adaptor_${adaptor}_all_in_one.yaml"
    sed "s#app.kubernetes.io/version: master#app.kubernetes.io/version: ${tag}#g" \
      "${CURR_DIR}/dist/octopus_adaptor_${adaptor}_all_in_one.yaml" >"${tmpfile}" && mv "${tmpfile}" "${CURR_DIR}/dist/octopus_adaptor_${adaptor}_all_in_one.yaml"
    sed "s#image: rancher/octopus-adaptor-${adaptor}:master#image: ${repo}/${image_name}:${tag}#g" \
      "${CURR_DIR}/dist/octopus_adaptor_${adaptor}_all_in_one.yaml" >"${tmpfile}" && mv "${tmpfile}" "${CURR_DIR}/dist/octopus_adaptor_${adaptor}_all_in_one.yaml"
  else
    octopus::log::warn "deploying manifest images has been stopped by WITHOUT_MANIFEST"
  fi

  octopus::log::info "...done"
}

function test() {
  [[ "${2:-}" != "only" ]] && build "$@"
  local adaptor="${1}"

  octopus::log::info "running unit tests for adaptor ${adaptor}..."

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
  [[ "${2:-}" != "only" ]] && test "$@"
  local adaptor="${1}"

  octopus::log::info "running integration tests for adaptor ${adaptor}..."

  CGO_ENABLED=0 go test \
    "${CURR_DIR}/test/integration/adaptor/..." -v -ginkgo.v -ginkgo.progress

  octopus::log::info "...done"
}

function e2e() {
  [[ "${2:-}" != "only" ]] && verify "$@"
  local adaptor="${1}"

  octopus::log::info "running E2E tests for adaptor ${adaptor}..."

  # TODO to implement the logic if needed, and place all E2E tests to test/e2e directory

  octopus::log::info "...done"
}

function entry() {
  local adaptor="${1:-}"
  shift 1

  local stage="${1:-build}"
  shift $(($# > 0 ? 1 : 0))

  octopus::log::info "make adaptor ${adaptor} ${stage} $*"

  case ${stage} in
  g | gen | generate) generate "${adaptor}" "$@" ;;
  m | mod) mod "${adaptor}" "$@" ;;
  l | lint) lint "${adaptor}" "$@" ;;
  b | build) build "${adaptor}" "$@" ;;
  p | pkg | package) package "${adaptor}" "$@" ;;
  d | dep | deploy) deploy "${adaptor}" "$@" ;;
  t | test) test "${adaptor}" "$@" ;;
  v | ver | verify) verify "${adaptor}" "$@" ;;
  e | e2e) e2e "${adaptor}" "$@" ;;
  *) octopus::log::fatal "unknown action '${stage}', select from generate,mod,lint,build,test,verify,package,deploy,e2e" ;;
  esac
}

if [[ ${BY:-} == "dapper" ]]; then
  octopus::dapper::run -C "${ROOT_DIR}" -f "adaptors/${1}/Dockerfile.dapper" "$@"
else
  entry "$@"
fi

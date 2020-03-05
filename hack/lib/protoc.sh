#!/usr/bin/env bash

function octopus::protoc::install_gen_gogofaster() {
  tmp_dir=$(mktemp -d)
  pushd "${tmp_dir}" >/dev/null || exist 1
  go mod init tmp
  go get github.com/gogo/protobuf/protoc-gen-gogofaster@v1.3.1
  rm -rf "${tmp_dir}"
  popd >/dev/null || return
}

function octopus::protoc::validate_gen_gogfaster() {
  if [[ -n "$(command -v protoc-gen-gogofaster)" ]]; then
    return 0
  fi

  octopus::log::info "installing protoc-gen-gogofaster"
  if octopus::protoc::install_gen_gogofaster; then
    export PATH="$PATH:${GOPATH}/bin/protoc-gen-gogofaster"
    octopus::log::info "installed protoc-gen-gogofaster"
    return 0
  fi
  octopus::log::error "no protoc-gen-gogofaster available"
  return 1
}

function octopus::protoc::validate() {
  if [[ -z "$(command -v protoc)" || "$(protoc --version)" != "libprotoc 3.11"* ]]; then
    octopus::log::error "generating protobuf requires protoc 3.11.0 or newer, please download and install the platform appropriate Protobuf package for your OS: https://github.com/protocolbuffers/protobuf/releases"
    return 1
  fi
  return 0
}

function octopus::protoc::protoc() {
  if ! octopus::protoc::validate; then
    octopus::log::fatal "protoc hasn't been installed"
  fi

  if ! octopus::protoc::validate_gen_gogfaster; then
    octopus::log::fatal "protoc-gen-gogfaster hasn't been installed"
  fi

  local pkg_path=${1}
  protoc \
    --proto_path="${pkg_path}" \
    --proto_path="${ROOT_DIR}/vendor" \
    --gogofaster_out=plugins=grpc:"${pkg_path}" \
    "${pkg_path}/api.proto"
}

function octopus::protoc::format() {
  local pkg_path=${1}
  cat "${ROOT_DIR}/hack/boilerplate.go.txt" "${pkg_path}/api.pb.go" >tmpfile && mv tmpfile "${pkg_path}/api.pb.go"
  go fmt "${pkg_path}/api.pb.go" >/dev/null 2>&1
}

function octopus::protoc::generate() {
  local pkg_path=${1}
  if [[ -f "${pkg_path}/api.proto" ]]; then
    octopus::protoc::protoc "${pkg_path}"
    octopus::protoc::format "${pkg_path}"
  fi
}

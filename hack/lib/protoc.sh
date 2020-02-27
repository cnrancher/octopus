#!/usr/bin/env bash

function octopus::protoc::validate_cli() {
  bin_path="$(command -v protoc)"
  if [[ -z "${bin_path}" || "$(protoc --version)" != "libprotoc 3.11"* ]]; then
    octopus::log::error "generating protobuf requires protoc 3.11.0 or newer, please download and install the platform appropriate Protobuf package for your OS: https://github.com/protocolbuffers/protobuf/releases"
    octopus::log::fatal "protobuf changes are not being validated"
  fi
  echo -n "${bin_path}"
}

function octopus::protoc::install_gen() {
  # use gogofaster
  if [[ -z "$(command -v protoc-gen-gogofaster)" ]]; then
    tmp_dir=$(mktemp -d)
    pushd "${tmp_dir}" || exist 1
    go mod init tmp
    go get github.com/gogo/protobuf/protoc-gen-gogofaster@v1.3.1
    rm -rf "${tmp_dir}"
    popd || return
  fi
}

function octopus::protoc:protoc() {
  gobin_path="$(go env GOBIN)"
  if [[ -z "${gobin_path}" ]]; then
    gobin_path="$(go env GOPATH)/bin"
  fi
  protoc_path=$(octopus::protoc::validate_cli)

  octopus::protoc::install_gen

  local pkg_path=${1}
  PATH="${gobin_path}:${PATH}" ${protoc_path} \
    --proto_path="${pkg_path}" \
    --proto_path="${ROOT_DIR}/vendor" \
    --gogofaster_out=plugins=grpc:"${pkg_path}" \
    "${pkg_path}/api.proto"
}

function octopus::protoc::format() {
  local pkg_path=${1}
  cat "${ROOT_DIR}/hack/boilerplate.go.txt" "${pkg_path}/api.pb.go" >tmpfile && mv tmpfile "${pkg_path}/api.pb.go"
  go fmt "${pkg_path}/api.pb.go"
}

function octopus::protoc::generate() {
  local pkg_path=${1}
  octopus::protoc:protoc "${pkg_path}"
  octopus::protoc::format "${pkg_path}"
}

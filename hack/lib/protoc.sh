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

function octopus::protoc::generate() {
  if ! octopus::protoc::validate_gen_gogfaster; then
    octopus::log::fatal "protoc-gen-gogfaster hasn't been installed"
  fi

  if ! octopus::protoc::validate; then
    octopus::log::error "cannot execute protoc as it hasn't installed"
    return
  fi

  local filepath="${1:-}"
  if [[ ! -f ${filepath} ]]; then
    octopus::log::warn "${filepath} isn't existed"
    return
  fi
  local filedir
  filedir=$(dirname "${filepath}")
  local filename
  filename=$(basename "${filepath}" ".proto")

  # generate
  protoc \
    --proto_path="${filedir}" \
    --proto_path="${ROOT_DIR}/vendor" \
    --gogofaster_out=plugins=grpc:"${filedir}" \
    "${filepath}"

  # format
  local tmpfile
  tmpfile=$(mktemp)
  local generated_file="${filedir}/${filename}.pb.go"
  sed "2d" "${generated_file}" >"${tmpfile}" && mv "${tmpfile}" "${generated_file}"
  cat "${ROOT_DIR}/hack/boilerplate.go.txt" "${generated_file}" >"${tmpfile}" && mv "${tmpfile}" "${generated_file}"
  go fmt "${generated_file}" >/dev/null 2>&1
}

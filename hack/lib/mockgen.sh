#!/usr/bin/env bash

function octopus::mockgen::install() {
  tmp_dir=$(mktemp -d)
  pushd "${tmp_dir}" >/dev/null || exit 1
  go mod init tmp
  go get github.com/golang/mock/mockgen@v1.4.3
  rm -rf "${tmp_dir}"
  popd >/dev/null || return
}

function octopus::mockgen::validate() {
  if [[ -n "$(command -v mockgen)" ]]; then
    return 0
  fi

  octopus::log::info "installing mockgen"
  if octopus::mockgen::install; then
    octopus::log::info "mockgen: $(mockgen --version)"
    return 0
  fi
  octopus::log::error "no mockgen available"
  return 1
}

function octopus::mockgen::generate_by_source() {
  if ! octopus::mockgen::validate; then
    octopus::log::error "cannot execute mockgen as it hasn't installed"
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
  filename=$(basename "${filepath}")
  local mocked_dir="${filedir}/mock"
  mkdir -p "${mocked_dir}"

  local mocked_file="${mocked_dir}/${filename}"
  # generate
  mockgen \
    -source="${filepath}" \
    -destination="${mocked_file}"

  # format
  local tmpfile
  tmpfile=$(mktemp)
  sed "2d" "${mocked_file}" >"${tmpfile}" && mv "${tmpfile}" "${mocked_file}"
  cat "${ROOT_DIR}/hack/boilerplate.go.txt" "${mocked_file}" >"${tmpfile}" && mv "${tmpfile}" "${mocked_file}"
  go fmt "${mocked_file}" >/dev/null 2>&1
}

function octopus::mockgen::generate_by_reflect() {
  if ! octopus::mockgen::validate; then
    octopus::log::error "cannot execute mockgen as it hasn't installed"
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
  filename=$(basename "${filepath}")
  local mocked_dir="${filedir}/mock"
  mkdir -p "${mocked_dir}"

  local mocked_file="${mocked_dir}/${filename}"
  pushd "${filedir}" >/dev/null 2>&1
  # generate
  GO111MODULE=on mockgen -destination="${mocked_file}" . "$(grep -e '^type\s.*interface\s{$' "${filepath}" | awk 'NR>1{printf ","} {printf $2}')"
  popd >/dev/null 2>&1

  # format
  local tmpfile
  tmpfile=$(mktemp)
  sed "2d" "${mocked_file}" >"${tmpfile}" && mv "${tmpfile}" "${mocked_file}"
  cat "${ROOT_DIR}/hack/boilerplate.go.txt" "${mocked_file}" >"${tmpfile}" && mv "${tmpfile}" "${mocked_file}"
  go fmt "${mocked_file}" >/dev/null 2>&1
}

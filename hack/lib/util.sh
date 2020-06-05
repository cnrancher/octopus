#!/usr/bin/env bash

function octopus::util::find_subdirs() {
  local path="$1"
  if [ -z "$path" ]; then
    path="./"
  fi
  # shellcheck disable=SC2010
  ls -l "$path" | grep "^d" | awk '{print $NF}'
}

function octopus::util::join_array() {
  local IFS="$1"
  shift 1
  echo "$*"
}

function octopus::util::get_os() {
  local os
  if go env GOOS >/dev/null 2>&1; then
    os=$(go env GOOS)
  else
    os=$(echo -n "$(uname -s)" | tr '[:upper:]' '[:lower:]')
  fi

  case ${os} in
  cygwin_nt*) os="windows" ;;
  mingw*) os="windows" ;;
  msys_nt*) os="windows" ;;
  esac

  echo -n "${os}"
}

function octopus::util::get_arch() {
  local arch
  if go env GOARCH >/dev/null 2>&1; then
    arch=$(go env GOARCH)
    if [[ "${arch}" == "arm" ]]; then
      arch="${arch}v$(go env GOARM)"
    fi
  else
    arch=$(uname -m)
  fi

  case ${arch} in
  armv5*) arch="armv5" ;;
  armv6*) arch="armv6" ;;
  armv7*)
    if [[ "${1:-}" == "--full-name" ]]; then
      arch="armv7"
    else
      arch="arm"
    fi
    ;;
  aarch64) arch="arm64" ;;
  x86) arch="386" ;;
  i686) arch="386" ;;
  i386) arch="386" ;;
  x86_64) arch="amd64" ;;
  esac

  echo -n "${arch}"
}

function octopus::util::get_random_port_start() {
  local offset="${1:-1}"
  if [[ ${offset} -le 0 ]]; then
    offset=1
  fi

  while true; do
    random_port=$((RANDOM % 10000 + 50000))
    for ((i = 0; i < offset; i++)); do
      if nc -z 127.0.0.1 $((random_port + i)); then
        random_port=0
        break
      fi
    done

    if [[ ${random_port} -ne 0 ]]; then
      echo "${random_port}"
      break
    fi
  done
}

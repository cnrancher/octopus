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
  shift
  echo "$*"
}

function octopus::util::get_os() {
  local os
  os=$(echo -n "$(uname -s)" | tr '[:upper:]' '[:lower:]')

  case ${os} in
  mingw*) os="windows" ;;
  esac

  echo -n "${os}"
}

function octopus::util::get_arch() {
  local arch
  arch=$(uname -m)

  case ${arch} in
  armv5*) arch="armv5" ;;
  armv6*) arch="armv6" ;;
  armv7*) arch="arm" ;;
  aarch64) arch="arm64" ;;
  x86) arch="386" ;;
  i686) arch="386" ;;
  i386) arch="386" ;;
  x86_64) arch="amd64" ;;
  amd64) arch="amd64" ;;
  esac

  echo -n "${arch}"
}

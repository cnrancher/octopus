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

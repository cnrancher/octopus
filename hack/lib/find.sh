#!/usr/bin/env bash

function octopus::find::subdirs() {
  local path="$1"
  if [ -z "$path" ]; then
    path="./"
  fi
  # shellcheck disable=SC2010
  ls -l "$path" | grep "^d" | awk '{print $NF}'
}

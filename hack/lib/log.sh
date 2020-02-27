#!/usr/bin/env bash

##
# Borrowed from github.com/kubernetes/kubernetes/hack/lib/logging.sh
##

# Handler for when we exit automatically on an error.
octopus::log::errexit() {
  local err="${PIPESTATUS[*]}"

  # if the shell we are in doesn't have errexit set (common in subshells) then
  # don't dump stacks.
  set +o | grep -qe "-o errexit" || return

  set +o xtrace
  octopus::log::fatal "${BASH_SOURCE[1]}:${BASH_LINENO[0]} '${BASH_COMMAND}' exited with status ${err}" "${1:-1}"
}

octopus::log::install_errexit() {
  # trap ERR to provide an error handler whenever a command exits nonzero, this
  # is a more verbose version of set -o errexit
  trap 'octopus::log::errexit' ERR

  # setting errtrace allows our ERR trap handler to be propagated to functions,
  # expansions and subshells
  set -o errtrace
}

# Info level logging.
octopus::log::info() {
  local message="${2:-}"

  local timestamp
  timestamp="$(date +"[%m%d %H:%M:%S]")"
  echo "[INFO] ${timestamp} ${1-}"
  shift
  for message; do
    echo "       ${message}"
  done
}

# Warn level logging.
octopus::log::warn() {
  local message="${1:-}"

  local timestamp
  timestamp="$(date +"[%m%d %H:%M:%S]")"
  echo "[WARN] ${timestamp} ${1-}"
  shift
  for message; do
    echo "       ${message}"
  done
}

# Error level logging, log an error but keep going, don't dump the stack or exit.
octopus::log::error() {
  local message="${1:-}"

  local timestamp
  timestamp="$(date +"[%m%d %H:%M:%S]")"
  echo "[ERRO] ${timestamp} ${1-}" >&2
  shift
  for message; do
    echo "       ${message}" >&2
  done
}

# Fatal level logging, dump the error stack and exit.
# Args:
#   $1 Message to log with the error
#   $2 The error code to return
#   $3 The number of stack frames to skip when printing.
octopus::log::fatal() {
  local message="${1:-}"
  local code="${2:-1}"

  local timestamp
  timestamp="$(date +"[%m%d %H:%M:%S]")"
  echo "[FATA] ${timestamp} ${message}" >&2

  # print out the stack trace described by $function_stack
  if [[ ${#FUNCNAME[@]} -gt 2 ]]; then
    echo "       call stack:" >&2
    local i
    for ((i = 1; i < ${#FUNCNAME[@]} - 2; i++)); do
      echo "       ${i}: ${BASH_SOURCE[${i} + 2]}:${BASH_LINENO[${i} + 1]} ${FUNCNAME[${i} + 1]}(...)" >&2
    done
  fi

  echo "[FATA] ${timestamp} exiting with status ${code}" >&2

  exit "${code}"
}

#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# The root of the octopus directory
ROOT_DIR="${CURR_DIR}"
source "${ROOT_DIR}/hack/lib/init.sh"

function entry() {
  # copy template
  read -p "Please input a camel-case/short name of adaptor (like 'Bluetooth', 'BLE'): " -r adaptorName
  if [[ -z "${adaptorName}" ]]; then
    octopus::log::fatal "the adaptor name is required"
  fi
  adaptorNameLowercase=$(echo -n "${adaptorName}" | tr '[:upper:]' '[:lower:]')
  adaptorPath="${ROOT_DIR}/adaptors/${adaptorNameLowercase}"
  if [[ -d "${adaptorPath}" ]]; then
    octopus::log::fatal "the directory is existed"
  fi
  cp -r "${ROOT_DIR}/template/adaptor" "${adaptorPath}"

  local tmpfile
  tmpfile=$(mktemp)

  # change api template to expected
  read -p "Please input a camel-case name of device (like 'BluetoothDevice'): " -r deviceName
  if [[ -z "${deviceName}" ]]; then
    octopus::log::fatal "the device name is required"
  fi
  deviceNameLowercase=$(echo -n "${deviceName}" | tr '[:upper:]' '[:lower:]')
  sed "s#TemplateDevice#${deviceName}#g" "${adaptorPath}/api/v1alpha1/templatedevice_types.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/api/v1alpha1/templatedevice_types.go"
  mv "${adaptorPath}/api/v1alpha1/templatedevice_types.go" "${adaptorPath}/api/v1alpha1/${deviceNameLowercase}_types.go"

  # change cmd template to expected
  sed "s#template/adaptor#adaptors/${adaptorNameLowercase}#g" "${adaptorPath}/cmd/template/main.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/cmd/template/main.go"
  sed "s#template#${adaptorNameLowercase}#g" "${adaptorPath}/cmd/template/main.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/cmd/template/main.go"
  mv "${adaptorPath}/cmd/template" "${adaptorPath}/cmd/${adaptorNameLowercase}"

  # change pkg template to expected
  sed "s#template/adaptor#adaptors/${adaptorNameLowercase}#g" "${adaptorPath}/pkg/adaptor/service.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/pkg/adaptor/service.go"
  sed "s#template/adaptor#adaptors/${adaptorNameLowercase}#g" "${adaptorPath}/pkg/template/template.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/pkg/template/template.go"
  sed "s#adaptors.edge.cattle.io/template#adaptors.edge.cattle.io/${adaptorNameLowercase}#g" "${adaptorPath}/pkg/template/template.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/pkg/template/template.go"
  sed "s#template.socket#${adaptorNameLowercase}.socket#g" "${adaptorPath}/pkg/template/template.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/pkg/template/template.go"
  sed "s#package template#package ${adaptorNameLowercase}#g" "${adaptorPath}/pkg/template/template.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/pkg/template/template.go"
  sed "s#templatedevices#${adaptorNameLowercase}devices#g" "${adaptorPath}/pkg/template/template.go" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/pkg/template/template.go"
  mv "${adaptorPath}/pkg/template/template.go" "${adaptorPath}/pkg/template/${adaptorNameLowercase}.go"
  mv "${adaptorPath}/pkg/template" "${adaptorPath}/pkg/${adaptorNameLowercase}"

  # change deploy template to expected
  mv "${adaptorPath}/deploy/manifests/crd/base/devices.edge.cattle.io_templatedevices.yaml" "${adaptorPath}/deploy/manifests/crd/base/devices.edge.cattle.io_${deviceNameLowercase}.yaml"
  sed "s#devices.edge.cattle.io_templatedevices.yaml#devices.edge.cattle.io_${deviceNameLowercase}.yaml#g" "${adaptorPath}/deploy/manifests/crd/kustomization.yaml" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/deploy/manifests/crd/kustomization.yaml"
  sed "s#octopus-adaptor-template#octopus-adaptor-${adaptorNameLowercase}#g" "${adaptorPath}/deploy/manifests/workload/daemonset.yaml" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/deploy/manifests/workload/daemonset.yaml"

  # change Dockerfile template to expected
  sed "s#template#${adaptorNameLowercase}#g" "${adaptorPath}/Dockerfile" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/Dockerfile"

  # change Dockerfile.dapper template to expected
  sed "s#template#${adaptorNameLowercase}#g" "${adaptorPath}/Dockerfile.dapper" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/Dockerfile.dapper"

  # change Makefile template to expected
  sed "s#template#${adaptorNameLowercase}#g" "${adaptorPath}/Makefile" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/Makefile"

  # change README.md template to expected
  sed "s#template#${adaptorNameLowercase}#g" "${adaptorPath}/README.md" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/README.md"
  sed "s#Template#${adaptorName}#g" "${adaptorPath}/README.md" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/README.md"
  sed "s#templatedevices#${deviceNameLowercase}s#g" "${adaptorPath}/README.md" >"${tmpfile}" && mv "${tmpfile}" "${adaptorPath}/README.md"

  # gofmt
  go fmt "${adaptorPath}/..." >/dev/null 2>&1

  # build
  make -se adaptor "${adaptorNameLowercase}" build
}

entry "$@"

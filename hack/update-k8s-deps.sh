#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on
export K8S_VERSION="${1:-v0.17.6}"

K8S_DEPS=(
  "k8s.io/api"
  "k8s.io/apiextensions-apiserver"
  "k8s.io/apimachinery"
  "k8s.io/apiserver"
  "k8s.io/code-generator"
  "k8s.io/client-go"
)

function update_module {
  local dep="${1}"
  local version="${2}"


  echo "Updating ${dep} to ${version}"

  go mod edit \
    -require="${dep}@${version}" \
    -replace="${dep}=${dep}@${version}"
}

for dep in "${K8S_DEPS[@]}"
do
  update_module "${dep}" "${K8S_VERSION}"
done


./hack/update-deps.sh


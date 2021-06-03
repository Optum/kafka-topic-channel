#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on

source $(dirname $0)/../vendor/knative.dev/hack/library.sh

readonly TMP_DIFFROOT="$(mktemp -d ${REPO_ROOT_DIR}/tmpdiffroot.XXXXXX)"

cleanup() {
  rm -rf "${TMP_DIFFROOT}"
}

trap "cleanup" EXIT SIGINT

cleanup

# Save working tree state
mkdir -p "${TMP_DIFFROOT}"

cp -aR \
  "${REPO_ROOT_DIR}/go.sum" \
  "${REPO_ROOT_DIR}/third_party" \
  "${REPO_ROOT_DIR}/vendor" \
  "${TMP_DIFFROOT}"

"${REPO_ROOT_DIR}/hack/update-codegen.sh"
echo "Diffing ${REPO_ROOT_DIR} against freshly generated codegen"
ret=0

diff -Naupr --no-dereference \
  "${REPO_ROOT_DIR}/third_party" "${TMP_DIFFROOT}/third_party" || ret=1

diff -Naupr --no-dereference \
  "${REPO_ROOT_DIR}/vendor" "${TMP_DIFFROOT}/vendor" || ret=1

# Restore working tree state
rm -fr \
  "${REPO_ROOT_DIR}/go.sum" \
  "${REPO_ROOT_DIR}/third_party" \
  "${REPO_ROOT_DIR}/vendor"

cp -aR "${TMP_DIFFROOT}"/* "${REPO_ROOT_DIR}"

if [[ $ret -eq 0 ]]
then
  echo "${REPO_ROOT_DIR} up to date."
 else
  echo "${REPO_ROOT_DIR} is out of date. Please run hack/update-codegen.sh"
  exit 1
fi

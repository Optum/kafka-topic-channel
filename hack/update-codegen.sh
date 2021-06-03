#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source $(dirname $0)/../vendor/knative.dev/hack/codegen-library.sh

# If we run with -mod=vendor here, then generate-groups.sh looks for vendor files in the wrong place.
export GOFLAGS=-mod=

echo "=== Update Codegen for $MODULE_NAME"

group "Kubernetes Codegen"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
"github.com/optum/kafka-topic-channel/pkg/client" "github.com/optum/kafka-topic-channel/pkg/apis" \
"messaging:v1alpha1" \
--go-header-file ${REPO_ROOT_DIR}/hack/boilerplate.go.txt

group "Knative Codegen"

# Knative Injection
${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
"github.com/optum/kafka-topic-channel/pkg/client" "github.com/optum/kafka-topic-channel/pkg/apis" \
"messaging:v1alpha1" \
--go-header-file ${REPO_ROOT_DIR}/hack/boilerplate.go.txt

group "Deepcopy Gen"

# Depends on generate-groups.sh to install bin/deepcopy-gen
${GOPATH}/bin/deepcopy-gen \
  -O zz_generated.deepcopy \
  --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate.go.txt

# save license
go-licenses save "./vendor/..." --save_path="./third_party/VENDOR-LICENSE/" --force

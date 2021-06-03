#!/usr/bin/env bash

# Documentation about this script and how to use it can be found
# at https://github.com/knative/test-infra/tree/master/ci

source $(dirname $0)/../vendor/knative.dev/hack/release.sh

export GO111MODULE=on

# Yaml files to generate, and the source config dir for them.
declare -A COMPONENTS
COMPONENTS=(
  ["channel-consolidated.yaml"]="config/channel/consolidated"
  ["channel-distributed.yaml"]="config/channel/distributed"
  ["source.yaml"]="config/source/single"
  ["mt-source.yaml"]="config/source/multi"
)
readonly COMPONENTS

function build_release() {
   # Update release labels if this is a tagged release
  if [[ -n "${TAG}" ]]; then
    echo "Tagged release, updating release labels to kafka.eventing.knative.dev/release: \"${TAG}\""
    LABEL_YAML_CMD=(sed -e "s|kafka.eventing.knative.dev/release: devel|kafka.eventing.knative.dev/release: \"${TAG}\"|")
  else
    echo "Untagged release, will NOT update release labels"
    LABEL_YAML_CMD=(cat)
  fi

  local all_yamls=()
  for yaml in "${!COMPONENTS[@]}"; do
    local config="${COMPONENTS[${yaml}]}"
    echo "Building Knative Eventing Contrib - ${config}"
    # TODO(chizhg): reenable --strict mode after https://github.com/knative/test-infra/issues/1262 is fixed.
    ko resolve ${KO_FLAGS} -f ${config}/ | "${LABEL_YAML_CMD[@]}" > ${yaml}
    all_yamls+=(${yaml})
  done

  if [ -d "${YAML_REPO_ROOT}/config/post-install" ]; then
    echo "Resolving post install manifests"

    local yaml="post-install.yaml"
    ko resolve ${KO_FLAGS} -f config/post-install/ | "${LABEL_YAML_CMD[@]}" > ${yaml}
    all_yamls+=(${yaml})
  fi

  ARTIFACTS_TO_PUBLISH="${all_yamls[@]}"
}

main $@

// +build e2e

package brokertrigger

import (
  "context"

  "knative.dev/reconciler-test/pkg/environment"
  "knative.dev/reconciler-test/pkg/feature"
  "knative.dev/reconciler-test/pkg/manifest"
)

func init() {
  environment.RegisterPackage(manifest.ImagesLocalYaml()...)
}

func Install() feature.StepFn {
  return func(ctx context.Context, t feature.T) {
    if _, err := manifest.InstallLocalYaml(ctx, nil); err != nil {
      t.Fatal(err)
    }
  }
}

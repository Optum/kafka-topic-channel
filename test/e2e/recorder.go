// +build e2e

package e2e

import (
  "time"

  "k8s.io/apimachinery/pkg/runtime/schema"
  "knative.dev/reconciler-test/pkg/eventshub"
  "knative.dev/reconciler-test/pkg/feature"
  "knative.dev/reconciler-test/pkg/k8s"
)

func RecorderFeature() *feature.Feature {
  svc := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

  to := "recorder"

  f := new(feature.Feature)

  f.Setup("install recorder", eventshub.Install(to, eventshub.StartReceiver))

  f.Requirement("recorder is addressable", k8s.IsAddressable(svc, to, time.Second, 30*time.Second))

  return f
}
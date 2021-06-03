// +build e2e

package e2e

import (
  "flag"
  "os"
  "time"
  "context"
  "testing"
  "strings"

  "knative.dev/pkg/injection"
  "knative.dev/pkg/system"

  // For our e2e testing, we want this linked first so that our
  // system namespace environment variable is defaulted prior to
  // logstream initialization.
  _ "github.com/optum/kafka-topic-channel/test/defaultsystem"
  "knative.dev/reconciler-test/pkg/environment"
  "knative.dev/reconciler-test/pkg/feature"
  "knative.dev/reconciler-test/pkg/k8s"
  "knative.dev/reconciler-test/pkg/knative"
)

func init() {
  environment.InitFlags(flag.CommandLine)
}

const (
  interval = 1 * time.Second
  timeout  = 5 * time.Minute
)

var global environment.GlobalEnvironment

func TestMain(m *testing.M) {
  flag.Parse()
  ctx, startInformers := injection.EnableInjectionOrDie(nil, nil) //nolint
  startInformers()
  global = environment.NewGlobalEnvironment(ctx)
  os.Exit(m.Run())
}

// TestChannelDirect makes sure a source delivers events to Sink.
func TestChannelDirect(t *testing.T) {
  t.Parallel()
  
  ctx, env := global.Environment(
    knative.WithKnativeNamespace(system.Namespace()),
    knative.WithLoggingConfig,
    knative.WithTracingConfig,
    k8s.WithEventListener,
  )
  env.Test(ctx, t, RecorderFeature())
  env.Test(ctx, t, DirectTest())
  env.Finish()
}

func AllGoReady(ctx context.Context, t feature.T) {
  env := environment.FromContext(ctx)
  for _, ref := range env.References() {
    if !strings.Contains(ref.APIVersion, "knative.dev") {
      // Let's not care so much about checking the status of non-Knative
      // resources.
      continue
    }    
    if err := k8s.WaitForReadyOrDone(ctx, ref, interval, timeout); err != nil {
      t.Fatal("failed to wait for ready or done, ", err, ref)
    }
  }
  t.Log("all resources ready")
}

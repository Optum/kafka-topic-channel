package main

import (
	"os"

	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"

	controller "github.com/optum/kafka-topic-channel/pkg/channel/kt/reconciler/dispatcher"
)

const component = "kafkachannel-dispatcher"

func main() {
	ctx := signals.NewContext()
	ns := os.Getenv("NAMESPACE")
	if ns != "" {
		ctx = injection.WithNamespaceScope(ctx, ns)
	}

	sharedmain.MainWithContext(ctx, component, controller.NewController)
}

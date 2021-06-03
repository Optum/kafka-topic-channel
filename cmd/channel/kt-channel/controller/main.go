package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/optum/kafka-topic-channel/pkg/channel/kt/reconciler/controller"
)

const component = "kafkatopicchannel-controller"

func main() {
	sharedmain.Main(component, controller.NewController)
}

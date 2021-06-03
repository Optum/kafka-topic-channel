// +build e2e

package e2e

import (
	"context"

	"github.com/optum/kafka-topic-channel/test/e2e/config/producer"
	"github.com/optum/kafka-topic-channel/test/e2e/config/brokertrigger"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
)

func DirectTest() *feature.Feature {
	f := new(feature.Feature)

	f.Setup("install broker trigger", brokertrigger.Install())
	f.Alpha("KTC channel").Must("goes ready", AllGoReady)
	f.Setup("install producer", producer.Install())
	f.Alpha("KTC channel").
		Must("the recorder received all sent events within the time",
			func(ctx context.Context, t feature.T) {
				eventshub.StoreFromContext(ctx, "recorder").AssertAtLeast(t, 5)
			})

	return f
}


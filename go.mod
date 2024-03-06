module github.com/optum/kafka-topic-channel

go 1.15

require (
	github.com/Shopify/sarama v1.28.0
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.4.1
	github.com/cloudevents/sdk-go/v2 v2.15.2
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.8.0
	go.opencensus.io v0.23.0
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.19.7
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
	knative.dev/eventing v0.22.1
	knative.dev/eventing-kafka v0.22.1-0.20210428064753-77c511a1c00d
	knative.dev/hack v0.0.0-20210428122153-93ad9129c268
	knative.dev/networking v0.0.0-20210428014353-796a80097840
	knative.dev/pkg v0.0.0-20210428141353-878c85083565
	knative.dev/reconciler-test v0.0.0-20210426151439-cd315b0f06aa
)

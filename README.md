# Kafka Topic Channel

[![Go Report Card](https://goreportcard.com/badge/github.com/optum/faas-swagger)](https://goreportcard.com/badge/github.com/optum/kafka-topic-channel)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/optum/kafka-topic-channel)
[![LICENSE](https://img.shields.io/github/license/knative/serving.svg)](https://github.com/optum/kafka-topic-channel/blob/master/LICENSE)

Kafka Topic Channel is a knative channel class based on kafka topic

### Why ?

There are already knative maintained channels like kafka consolidated, kafka distributed, inmemory channel, why new channel? The answer is 

Knative eventing primitives like broker, trigger and subscription are great for event processing, but they rely on the underlying channel. The existing kafka based channels needs kafka admin rights and requires the knative operator to maintain the kafka cluster or find a managed kafka cluster with admin rights. As a knative operator its challenging to maintain another big solution like kafka. If you are in the same boat, this kafka channel solution might work for you.

This implementation,

* Creates a knative channel based on the Kafka Topic information in any namespace.
* Kafka connection information is identical to kafka source
* Other knative primitives like broker, trigger and subscription works as is on this channel
* Each kafka topic channel resource would create a separate dispatcher/channel pod
* Bring your own kafka topic, get a knative channel

### Installation

Prereq for using this channel is knative eventing with broker. 

Once you have knative eventing in your cluster, install knative-topic-channel.

```shell
kubectl apply -f config/release/release.yaml
```

### Usage

Please take a look at the [samples](./samples) on how to use this channel.

### Contributing to the Project
The team is open to contributions to our project. For more details, see our [Contribution Guide.](./CONTRIBUTING.md)

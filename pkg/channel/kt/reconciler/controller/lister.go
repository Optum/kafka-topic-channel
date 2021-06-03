package controller

import (
	"context"
	"fmt"
	"net/url"

	"github.com/optum/kafka-topic-channel/pkg/channel/kt/status"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
	v1 "k8s.io/client-go/listers/core/v1"
)

type DispatcherPodsLister struct {
	logger         *zap.SugaredLogger
	endpointLister v1.EndpointsLister
}

func (t *DispatcherPodsLister) ListProbeTargets(ctx context.Context, kc v1alpha1.KafkaTopicChannel) (*status.ProbeTarget, error) {
	dispatcherNamespace := kc.Namespace

	// Get the Dispatcher Service Endpoints and propagate the status to the Channel
	// endpoints has the same name as the service, so not a bug.
	eps, err := t.endpointLister.Endpoints(dispatcherNamespace).Get(fmt.Sprintf("%s-dispatcher", kc.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to get internal service: %w", err)
	}
	var readyIPs []string

	for _, sub := range eps.Subsets {
		for _, address := range sub.Addresses {
			readyIPs = append(readyIPs, address.IP)
		}
	}

	if len(readyIPs) == 0 {
		return nil, fmt.Errorf("no gateway pods available")
	}

	u, _ := url.Parse(fmt.Sprintf("http://%s.%s/%s/%s", fmt.Sprintf("%s-dispatcher", kc.Name), dispatcherNamespace, kc.Namespace, kc.Name))

	return &status.ProbeTarget{
		PodIPs:  sets.NewString(readyIPs...),
		PodPort: "8081",
		URL:     u,
	}, nil
}

func NewProbeTargetLister(logger *zap.SugaredLogger, lister v1.EndpointsLister) status.ProbeTargetLister {
	tl := DispatcherPodsLister{
		logger:         logger,
		endpointLister: lister,
	}
	return &tl
}

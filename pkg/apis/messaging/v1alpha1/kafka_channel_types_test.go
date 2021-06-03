package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	eventingduck "knative.dev/eventing/pkg/apis/duck/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestKafkaTopicChannel_GetGroupVersionKind(t *testing.T) {
	ch := KafkaTopicChannel{}
	gvk := ch.GetGroupVersionKind()

	if gvk.Kind != "KafkaTopicChannel" {
		t.Errorf("Should be 'KafkaTopicChannel'.")
	}
}

func TestKafkaTopicChannelGetStatus(t *testing.T) {
	status := &duckv1.Status{}
	config := KafkaTopicChannel{
		Status: KafkaTopicChannelStatus{
			ChannelableStatus: eventingduck.ChannelableStatus{
				Status: *status,
			},
		},
	}

	if !cmp.Equal(config.GetStatus(), status) {
		t.Errorf("GetStatus did not retrieve status. Got=%v Want=%v", config.GetStatus(), status)
	}
}

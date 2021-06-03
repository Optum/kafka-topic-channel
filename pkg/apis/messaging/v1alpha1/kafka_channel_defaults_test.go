package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing-kafka/pkg/common/constants"
)

const (
	testNumPartitions     = 10
	testReplicationFactor = 5
)

func TestKafkaTopicChannelDefaults(t *testing.T) {
	testCases := map[string]struct {
		initial  KafkaTopicChannel
		expected KafkaTopicChannel
	}{
		"nil spec": {
			initial: KafkaTopicChannel{},
			expected: KafkaTopicChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"messaging.knative.dev/subscribable": "v1"},
				},
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     constants.DefaultNumPartitions,
					ReplicationFactor: constants.DefaultReplicationFactor,
				},
			},
		},
		"numPartitions not set": {
			initial: KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{
					ReplicationFactor: testReplicationFactor,
				},
			},
			expected: KafkaTopicChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"messaging.knative.dev/subscribable": "v1"},
				},
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     constants.DefaultNumPartitions,
					ReplicationFactor: testReplicationFactor,
				},
			},
		},
		"replicationFactor not set": {
			initial: KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{
					NumPartitions: testNumPartitions,
				},
			},
			expected: KafkaTopicChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"messaging.knative.dev/subscribable": "v1"},
				},
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     testNumPartitions,
					ReplicationFactor: constants.DefaultReplicationFactor,
				},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			tc.initial.SetDefaults(context.TODO())
			if diff := cmp.Diff(tc.expected, tc.initial); diff != "" {
				t.Fatalf("Unexpected defaults (-want, +got): %s", diff)
			}
		})
	}
}

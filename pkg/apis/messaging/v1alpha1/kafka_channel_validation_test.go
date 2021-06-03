package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/webhook/resourcesemantics"

	eventingduck "knative.dev/eventing/pkg/apis/duck/v1"
	"knative.dev/eventing/pkg/apis/eventing"
	"knative.dev/pkg/apis"
)

func TestKafkaTopicChannelValidation(t *testing.T) {

	testCases := map[string]struct {
		cr   resourcesemantics.GenericCRD
		want *apis.FieldError
	}{
		"empty spec": {
			cr: &KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{},
			},
			want: func() *apis.FieldError {
				var errs *apis.FieldError
				fe := apis.ErrInvalidValue(0, "spec.numPartitions")
				errs = errs.Also(fe)
				fe = apis.ErrInvalidValue(0, "spec.replicationFactor")
				errs = errs.Also(fe)
				return errs
			}(),
		},
		"negative numPartitions": {
			cr: &KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     -10,
					ReplicationFactor: 1,
				},
			},
			want: func() *apis.FieldError {
				fe := apis.ErrInvalidValue(-10, "spec.numPartitions")
				return fe
			}(),
		},
		"negative replicationFactor": {
			cr: &KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     1,
					ReplicationFactor: -10,
				},
			},
			want: func() *apis.FieldError {
				fe := apis.ErrInvalidValue(-10, "spec.replicationFactor")
				return fe
			}(),
		},
		"valid subscribers array": {
			cr: &KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     1,
					ReplicationFactor: 1,
					ChannelableSpec: eventingduck.ChannelableSpec{
						SubscribableSpec: eventingduck.SubscribableSpec{
							Subscribers: []eventingduck.SubscriberSpec{{
								SubscriberURI: apis.HTTP("subscriberendpoint"),
								ReplyURI:      apis.HTTP("resultendpoint"),
							}},
						}},
				},
			},
			want: nil,
		},
		"empty subscriber at index 1": {
			cr: &KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     1,
					ReplicationFactor: 1,
					ChannelableSpec: eventingduck.ChannelableSpec{
						SubscribableSpec: eventingduck.SubscribableSpec{
							Subscribers: []eventingduck.SubscriberSpec{{
								SubscriberURI: apis.HTTP("subscriberendpoint"),
								ReplyURI:      apis.HTTP("replyendpoint"),
							}, {}},
						}},
				},
			},
			want: func() *apis.FieldError {
				fe := apis.ErrMissingField("spec.subscribable.subscriber[1].replyURI", "spec.subscribable.subscriber[1].subscriberURI")
				fe.Details = "expected at least one of, got none"
				return fe
			}(),
		},
		"two empty subscribers": {
			cr: &KafkaTopicChannel{
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     1,
					ReplicationFactor: 1,
					ChannelableSpec: eventingduck.ChannelableSpec{
						SubscribableSpec: eventingduck.SubscribableSpec{
							Subscribers: []eventingduck.SubscriberSpec{{}, {}},
						},
					},
				},
			},
			want: func() *apis.FieldError {
				var errs *apis.FieldError
				fe := apis.ErrMissingField("spec.subscribable.subscriber[0].replyURI", "spec.subscribable.subscriber[0].subscriberURI")
				fe.Details = "expected at least one of, got none"
				errs = errs.Also(fe)
				fe = apis.ErrMissingField("spec.subscribable.subscriber[1].replyURI", "spec.subscribable.subscriber[1].subscriberURI")
				fe.Details = "expected at least one of, got none"
				errs = errs.Also(fe)
				return errs
			}(),
		},
		"invalid scope annotation": {
			cr: &KafkaTopicChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						eventing.ScopeAnnotationKey: "notvalid",
					},
				},
				Spec: KafkaTopicChannelSpec{
					NumPartitions:     1,
					ReplicationFactor: 1,
				},
			},
			want: func() *apis.FieldError {
				fe := apis.ErrInvalidValue("notvalid", "metadata.annotations.[eventing.knative.dev/scope]")
				fe.Details = "expected either 'cluster' or 'namespace'"
				return fe
			}(),
		},
	}

	for n, test := range testCases {
		t.Run(n, func(t *testing.T) {
			got := test.cr.Validate(context.Background())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("%s: validate (-want, +got) = %v", n, diff)
			}
		})
	}
}

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"knative.dev/pkg/apis"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var condReady = apis.Condition{
	Type:   KafkaTopicChannelConditionReady,
	Status: corev1.ConditionTrue,
}

var condDispatcherNotReady = apis.Condition{
	Type:   KafkaTopicChannelConditionDispatcherReady,
	Status: corev1.ConditionFalse,
}

var deploymentConditionReady = appsv1.DeploymentCondition{
	Type:   appsv1.DeploymentAvailable,
	Status: corev1.ConditionTrue,
}

var deploymentConditionNotReady = appsv1.DeploymentCondition{
	Type:   appsv1.DeploymentAvailable,
	Status: corev1.ConditionFalse,
}

var deploymentStatusReady = &appsv1.DeploymentStatus{Conditions: []appsv1.DeploymentCondition{deploymentConditionReady}}
var deploymentStatusNotReady = &appsv1.DeploymentStatus{Conditions: []appsv1.DeploymentCondition{deploymentConditionNotReady}}

var ignoreAllButTypeAndStatus = cmpopts.IgnoreFields(
	apis.Condition{},
	"LastTransitionTime", "Message", "Reason", "Severity")

func TestKafkaTopicChannelGetConditionSet(t *testing.T) {
	r := &KafkaTopicChannel{}

	if got, want := r.GetConditionSet().GetTopLevelConditionType(), apis.ConditionReady; got != want {
		t.Errorf("GetTopLevelCondition=%v, want=%v", got, want)
	}
}

func TestChannelGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		cs        *KafkaTopicChannelStatus
		condQuery apis.ConditionType
		want      *apis.Condition
	}{{
		name: "single condition",
		cs: &KafkaTopicChannelStatus{
			ChannelableStatus: eventingduckv1.ChannelableStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						condReady,
					},
				},
			},
		},
		condQuery: apis.ConditionReady,
		want:      &condReady,
	}, {
		name: "unknown condition",
		cs: &KafkaTopicChannelStatus{
			ChannelableStatus: eventingduckv1.ChannelableStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						condReady,
						condDispatcherNotReady,
					},
				},
			},
		},
		condQuery: apis.ConditionType("foo"),
		want:      nil,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.cs.GetCondition(test.condQuery)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected condition (-want, +got) = %v", diff)
			}
		})
	}
}

func TestChannelInitializeConditions(t *testing.T) {
	tests := []struct {
		name string
		cs   *KafkaTopicChannelStatus
		want *KafkaTopicChannelStatus
	}{{
		name: "empty",
		cs:   &KafkaTopicChannelStatus{},
		want: &KafkaTopicChannelStatus{
			ChannelableStatus: eventingduckv1.ChannelableStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   KafkaTopicChannelConditionAddressable,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionChannelServiceReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionConfigReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionDispatcherReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionEndpointsReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionServiceReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionTopicReady,
						Status: corev1.ConditionUnknown,
					}},
				},
			},
		},
	}, {
		name: "one false",
		cs: &KafkaTopicChannelStatus{
			ChannelableStatus: eventingduckv1.ChannelableStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   KafkaTopicChannelConditionDispatcherReady,
						Status: corev1.ConditionFalse,
					}},
				},
			},
		},
		want: &KafkaTopicChannelStatus{
			ChannelableStatus: eventingduckv1.ChannelableStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   KafkaTopicChannelConditionAddressable,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionChannelServiceReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionConfigReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionDispatcherReady,
						Status: corev1.ConditionFalse,
					}, {
						Type:   KafkaTopicChannelConditionEndpointsReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionServiceReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionTopicReady,
						Status: corev1.ConditionUnknown,
					}},
				},
			},
		},
	}, {
		name: "one true",
		cs: &KafkaTopicChannelStatus{
			ChannelableStatus: eventingduckv1.ChannelableStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   KafkaTopicChannelConditionDispatcherReady,
						Status: corev1.ConditionTrue,
					}},
				},
			},
		},
		want: &KafkaTopicChannelStatus{
			ChannelableStatus: eventingduckv1.ChannelableStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   KafkaTopicChannelConditionAddressable,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionChannelServiceReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionConfigReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionDispatcherReady,
						Status: corev1.ConditionTrue,
					}, {
						Type:   KafkaTopicChannelConditionEndpointsReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionServiceReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   KafkaTopicChannelConditionTopicReady,
						Status: corev1.ConditionUnknown,
					}},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.cs.InitializeConditions()
			if diff := cmp.Diff(test.want, test.cs, ignoreAllButTypeAndStatus); diff != "" {
				t.Errorf("unexpected conditions (-want, +got) = %v", diff)
			}
		})
	}
}

func TestChannelIsReady(t *testing.T) {
	tests := []struct {
		name                    string
		markServiceReady        bool
		markChannelServiceReady bool
		markConfigurationReady  bool
		setAddress              bool
		markEndpointsReady      bool
		markTopicReady          bool
		wantReady               bool
		dispatcherStatus        *appsv1.DeploymentStatus
	}{{
		name:                    "all happy",
		markServiceReady:        true,
		markChannelServiceReady: true,
		markConfigurationReady:  true,
		markEndpointsReady:      true,
		dispatcherStatus:        deploymentStatusReady,
		setAddress:              true,
		markTopicReady:          true,
		wantReady:               true,
	}, {
		name:                    "service not ready",
		markServiceReady:        false,
		markChannelServiceReady: false,
		markConfigurationReady:  true,
		markEndpointsReady:      true,
		dispatcherStatus:        deploymentStatusReady,
		setAddress:              true,
		markTopicReady:          true,
		wantReady:               false,
	}, {
		name:                    "endpoints not ready",
		markServiceReady:        true,
		markChannelServiceReady: false,
		markConfigurationReady:  true,
		markEndpointsReady:      false,
		dispatcherStatus:        deploymentStatusReady,
		setAddress:              true,
		markTopicReady:          true,
		wantReady:               false,
	}, {
		name:                    "deployment not ready",
		markServiceReady:        true,
		markConfigurationReady:  true,
		markEndpointsReady:      true,
		markChannelServiceReady: false,
		dispatcherStatus:        deploymentStatusNotReady,
		setAddress:              true,
		markTopicReady:          true,
		wantReady:               false,
	}, {
		name:                    "address not set",
		markServiceReady:        true,
		markConfigurationReady:  true,
		markChannelServiceReady: false,
		markEndpointsReady:      true,
		dispatcherStatus:        deploymentStatusReady,
		setAddress:              false,
		markTopicReady:          true,
		wantReady:               false,
	}, {
		name:                    "channel service not ready",
		markServiceReady:        true,
		markConfigurationReady:  true,
		markChannelServiceReady: false,
		markEndpointsReady:      true,
		dispatcherStatus:        deploymentStatusReady,
		setAddress:              true,
		markTopicReady:          true,
		wantReady:               false,
	}, {
		name:                    "topic not ready",
		markServiceReady:        true,
		markConfigurationReady:  true,
		markChannelServiceReady: true,
		markEndpointsReady:      true,
		dispatcherStatus:        deploymentStatusReady,
		setAddress:              true,
		markTopicReady:          false,
		wantReady:               false,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cs := &KafkaTopicChannelStatus{}
			cs.InitializeConditions()
			if test.markServiceReady {
				cs.MarkServiceTrue()
			} else {
				cs.MarkServiceFailed("NotReadyService", "testing")
			}
			if test.markChannelServiceReady {
				cs.MarkChannelServiceTrue()
			} else {
				cs.MarkChannelServiceFailed("NotReadyChannelService", "testing")
			}
			if test.markConfigurationReady {
				cs.MarkConfigTrue()
			} else {
				cs.MarkConfigFailed("NotReadyConfiguration", "testing")
			}
			if test.setAddress {
				cs.SetAddress(&apis.URL{Scheme: "http", Host: "foo.bar"})
			}
			if test.markEndpointsReady {
				cs.MarkEndpointsTrue()
			} else {
				cs.MarkEndpointsFailed("NotReadyEndpoints", "testing")
			}
			if test.dispatcherStatus != nil {
				cs.PropagateDispatcherStatus(test.dispatcherStatus)
			} else {
				cs.MarkDispatcherFailed("NotReadyDispatcher", "testing")
			}
			if test.markTopicReady {
				cs.MarkTopicTrue()
			} else {
				cs.MarkTopicFailed("NotReadyTopic", "testing")
			}
			got := cs.IsReady()
			if test.wantReady != got {
				t.Errorf("unexpected readiness: want %v, got %v", test.wantReady, got)
			}
		})
	}
}

func TestKafkaTopicChannelStatus_SetAddressable(t *testing.T) {
	testCases := map[string]struct {
		url  *apis.URL
		want *KafkaTopicChannelStatus
	}{
		"empty string": {
			want: &KafkaTopicChannelStatus{
				ChannelableStatus: eventingduckv1.ChannelableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{
							{
								Type:   KafkaTopicChannelConditionAddressable,
								Status: corev1.ConditionFalse,
							},
							// Note that Ready is here because when the condition is marked False, duck
							// automatically sets Ready to false.
							{
								Type:   KafkaTopicChannelConditionReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
					AddressStatus: duckv1.AddressStatus{Address: &duckv1.Addressable{}},
				},
			},
		},
		"has domain": {
			url: &apis.URL{Scheme: "http", Host: "test-domain"},
			want: &KafkaTopicChannelStatus{
				ChannelableStatus: eventingduckv1.ChannelableStatus{
					AddressStatus: duckv1.AddressStatus{
						Address: &duckv1.Addressable{
							URL: &apis.URL{
								Scheme: "http",
								Host:   "test-domain",
							},
						},
					},
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   KafkaTopicChannelConditionAddressable,
							Status: corev1.ConditionTrue,
						}, {
							// Ready unknown comes from other dependent conditions via MarkTrue.
							Type:   KafkaTopicChannelConditionReady,
							Status: corev1.ConditionUnknown,
						}},
					},
				},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			cs := &KafkaTopicChannelStatus{}
			cs.SetAddress(tc.url)
			if diff := cmp.Diff(tc.want, cs, ignoreAllButTypeAndStatus); diff != "" {
				t.Errorf("unexpected conditions (-want, +got) = %v", diff)
			}
		})
	}
}

func TestRegisterAlternateKafkaTopicChannelConditionSet(t *testing.T) {

	cs := apis.NewLivingConditionSet(apis.ConditionReady, "hello")

	RegisterAlternateKafkaTopicChannelConditionSet(cs)

	kc := KafkaTopicChannel{}

	assert.Equal(t, cs, kc.GetConditionSet())
	assert.Equal(t, cs, kc.Status.GetConditionSet())
}

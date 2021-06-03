package v1alpha1

import (
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var kc = apis.NewLivingConditionSet(
	KafkaTopicChannelConditionTopicReady,
	KafkaTopicChannelConditionDispatcherReady,
	KafkaTopicChannelConditionServiceReady,
	KafkaTopicChannelConditionEndpointsReady,
	KafkaTopicChannelConditionAddressable,
	KafkaTopicChannelConditionChannelServiceReady,
	KafkaTopicChannelConditionConfigReady)
var channelCondSetLock = sync.RWMutex{}

const (
	// KafkaTopicChannelConditionReady has status True when all subconditions below have been set to True.
	KafkaTopicChannelConditionReady = apis.ConditionReady

	// KafkaTopicChannelConditionDispatcherReady has status True when a Dispatcher deployment is ready
	// Keyed off appsv1.DeploymentAvailable, which means minimum available replicas required are up
	// and running for at least minReadySeconds.
	KafkaTopicChannelConditionDispatcherReady apis.ConditionType = "DispatcherReady"

	// KafkaTopicChannelConditionServiceReady has status True when a k8s Service is ready. This
	// basically just means it exists because there's no meaningful status in Service. See Endpoints
	// below.
	KafkaTopicChannelConditionServiceReady apis.ConditionType = "ServiceReady"

	// KafkaTopicChannelConditionEndpointsReady has status True when a k8s Service Endpoints are backed
	// by at least one endpoint.
	KafkaTopicChannelConditionEndpointsReady apis.ConditionType = "EndpointsReady"

	// KafkaTopicChannelConditionAddressable has status true when this KafkaTopicChannel meets
	// the Addressable contract and has a non-empty URL.
	KafkaTopicChannelConditionAddressable apis.ConditionType = "Addressable"

	// KafkaTopicChannelConditionServiceReady has status True when a k8s Service representing the channel is ready.
	// Because this uses ExternalName, there are no endpoints to check.
	KafkaTopicChannelConditionChannelServiceReady apis.ConditionType = "ChannelServiceReady"

	// KafkaTopicChannelConditionTopicReady has status True when the Kafka topic to use by the channel exists.
	KafkaTopicChannelConditionTopicReady apis.ConditionType = "TopicReady"

	// KafkaTopicChannelConditionConfigReady has status True when the Kafka configuration to use by the channel exists and is valid
	// (ie. the connection has been established).
	KafkaTopicChannelConditionConfigReady apis.ConditionType = "ConfigurationReady"
)

// RegisterAlternateKafkaTopicChannelConditionSet register a different apis.ConditionSet.
func RegisterAlternateKafkaTopicChannelConditionSet(conditionSet apis.ConditionSet) {
	channelCondSetLock.Lock()
	defer channelCondSetLock.Unlock()

	kc = conditionSet
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*KafkaTopicChannel) GetConditionSet() apis.ConditionSet {
	channelCondSetLock.RLock()
	defer channelCondSetLock.RUnlock()

	return kc
}

// GetConditionSet retrieves the condition set for this resource.
func (*KafkaTopicChannelStatus) GetConditionSet() apis.ConditionSet {
	channelCondSetLock.RLock()
	defer channelCondSetLock.RUnlock()

	return kc
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (cs *KafkaTopicChannelStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return cs.GetConditionSet().Manage(cs).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (cs *KafkaTopicChannelStatus) IsReady() bool {
	return cs.GetConditionSet().Manage(cs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (cs *KafkaTopicChannelStatus) InitializeConditions() {
	cs.GetConditionSet().Manage(cs).InitializeConditions()
}

// SetAddress sets the address (as part of Addressable contract) and marks the correct condition.
func (cs *KafkaTopicChannelStatus) SetAddress(url *apis.URL) {
	if cs.Address == nil {
		cs.Address = &duckv1.Addressable{}
	}
	if url != nil {
		cs.Address.URL = url
		cs.GetConditionSet().Manage(cs).MarkTrue(KafkaTopicChannelConditionAddressable)
	} else {
		cs.Address.URL = nil
		cs.GetConditionSet().Manage(cs).MarkFalse(KafkaTopicChannelConditionAddressable, "EmptyURL", "URL is nil")
	}
}

func (cs *KafkaTopicChannelStatus) MarkDispatcherFailed(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkFalse(KafkaTopicChannelConditionDispatcherReady, reason, messageFormat, messageA...)
}

func (cs *KafkaTopicChannelStatus) MarkDispatcherUnknown(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkUnknown(KafkaTopicChannelConditionDispatcherReady, reason, messageFormat, messageA...)
}

// TODO: Unify this with the ones from Eventing. Say: Broker, Trigger.
func (cs *KafkaTopicChannelStatus) PropagateDispatcherStatus(ds *appsv1.DeploymentStatus) {
	for _, cond := range ds.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			if cond.Status == corev1.ConditionTrue {
				cs.GetConditionSet().Manage(cs).MarkTrue(KafkaTopicChannelConditionDispatcherReady)
			} else if cond.Status == corev1.ConditionFalse {
				cs.MarkDispatcherFailed("DispatcherDeploymentFalse", "The status of Dispatcher Deployment is False: %s : %s", cond.Reason, cond.Message)
			} else if cond.Status == corev1.ConditionUnknown {
				cs.MarkDispatcherUnknown("DispatcherDeploymentUnknown", "The status of Dispatcher Deployment is Unknown: %s : %s", cond.Reason, cond.Message)
			}
		}
	}
}

func (cs *KafkaTopicChannelStatus) MarkServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkFalse(KafkaTopicChannelConditionServiceReady, reason, messageFormat, messageA...)
}

func (cs *KafkaTopicChannelStatus) MarkServiceUnknown(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkUnknown(KafkaTopicChannelConditionServiceReady, reason, messageFormat, messageA...)
}

func (cs *KafkaTopicChannelStatus) MarkServiceTrue() {
	cs.GetConditionSet().Manage(cs).MarkTrue(KafkaTopicChannelConditionServiceReady)
}

func (cs *KafkaTopicChannelStatus) MarkChannelServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkFalse(KafkaTopicChannelConditionChannelServiceReady, reason, messageFormat, messageA...)
}

func (cs *KafkaTopicChannelStatus) MarkChannelServiceTrue() {
	cs.GetConditionSet().Manage(cs).MarkTrue(KafkaTopicChannelConditionChannelServiceReady)
}

func (cs *KafkaTopicChannelStatus) MarkEndpointsFailed(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkFalse(KafkaTopicChannelConditionEndpointsReady, reason, messageFormat, messageA...)
}

func (cs *KafkaTopicChannelStatus) MarkEndpointsTrue() {
	cs.GetConditionSet().Manage(cs).MarkTrue(KafkaTopicChannelConditionEndpointsReady)
}

func (cs *KafkaTopicChannelStatus) MarkTopicTrue() {
	cs.GetConditionSet().Manage(cs).MarkTrue(KafkaTopicChannelConditionTopicReady)
}

func (cs *KafkaTopicChannelStatus) MarkTopicFailed(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkFalse(KafkaTopicChannelConditionTopicReady, reason, messageFormat, messageA...)
}

func (cs *KafkaTopicChannelStatus) MarkConfigTrue() {
	cs.GetConditionSet().Manage(cs).MarkTrue(KafkaTopicChannelConditionConfigReady)
}

func (cs *KafkaTopicChannelStatus) MarkConfigFailed(reason, messageFormat string, messageA ...interface{}) {
	cs.GetConditionSet().Manage(cs).MarkFalse(KafkaTopicChannelConditionConfigReady, reason, messageFormat, messageA...)
}

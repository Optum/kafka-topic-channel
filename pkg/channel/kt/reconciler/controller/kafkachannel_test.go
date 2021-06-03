package controller

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	"github.com/optum/kafka-topic-channel/pkg/channel/kt/reconciler/controller/resources"
	reconcilertesting "github.com/optum/kafka-topic-channel/pkg/channel/kt/reconciler/testing"
	fakekafkaclient "github.com/optum/kafka-topic-channel/pkg/client/injection/client/fake"
	kafkachannel "github.com/optum/kafka-topic-channel/pkg/client/injection/reconciler/messaging/v1alpha1/kafkatopicchannel"
	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	eventingClient "knative.dev/eventing/pkg/client/injection/client"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	. "knative.dev/pkg/reconciler/testing"
)

const (
	testNS                       = "test-namespace"
	testDispatcherserviceAccount = "kt-channel-dispatcher"
	testEventConfigReadCRName    = "kt-channel-dispatcher-test-namespace"
	testDispatcherCR             = "kt-channel-dispatcher"
	kcName                       = "test-kc"
	testDispatcherName           = "test-kc-dispatcher"
	testDispatcherImage          = "test-image"
	channelServiceAddress        = "test-kc-kn-channel.test-namespace.svc.cluster.local"
	brokerName                   = "test-broker"
	finalizerName                = "kafkatopicchannels.messaging.optum.dev"
	sub1UID                      = "2f9b5e8e-deb6-11e8-9f32-f2801f1b9fd1"
	sub2UID                      = "34c5aec8-deb6-11e8-9f32-f2801f1b9fd1"
	twoSubscribersPatch          = `[{"op":"add","path":"/status/subscribers","value":[{"observedGeneration":1,"ready":"True","uid":"2f9b5e8e-deb6-11e8-9f32-f2801f1b9fd1"},{"observedGeneration":2,"ready":"True","uid":"34c5aec8-deb6-11e8-9f32-f2801f1b9fd1"}]}]`
)

var (
	finalizerUpdatedEvent = Eventf(corev1.EventTypeNormal, "FinalizerUpdate", `Updated "test-kc" finalizers`)
)

func init() {
	// Add types to scheme
	_ = v1alpha1.AddToScheme(scheme.Scheme)
	_ = duckv1.AddToScheme(scheme.Scheme)
}

func TestAllCases(t *testing.T) {
	kcKey := testNS + "/" + kcName
	table := TableTest{
		{
			Name: "bad workqueue key",
			// Make sure Reconcile handles bad keys.
			Key: "too/many/parts",
		}, {
			Name: "key not found",
			// Make sure Reconcile handles good keys that don't exist.
			Key: "foo/not-found",
		}, {
			Name: "deleting",
			Key:  kcKey,
			Objects: []runtime.Object{
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaChannelDeleted)},
			WantErr: false,
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", `KafkaTopicChannel reconciled: "test-namespace/test-kc"`),
			},
		}, {
			Name: "deployment does not exist, automatically created and patching finalizers",
			Key:  kcKey,
			Objects: []runtime.Object{
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaChannelTopicReady()),
			},
			WantErr: true,
			WantCreates: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsNotReady("DispatcherEndpointsDoesNotExist", "Dispatcher Endpoints does not exist")),
			}},
			WantPatches: []clientgotesting.PatchActionImpl{
				patchFinalizers(testNS, kcName),
			},
			WantEvents: []string{
				finalizerUpdatedEvent,
				Eventf(corev1.EventTypeNormal, dispatcherServiceAccountCreated, "Dispatcher service account created"),
				Eventf(corev1.EventTypeNormal, dispatcherRoleBindingCreated, "Dispatcher role binding created"),
				Eventf(corev1.EventTypeNormal, dispatcherRoleBindingCreated, "Dispatcher role binding created"),
				Eventf(corev1.EventTypeNormal, dispatcherDeploymentCreated, "Dispatcher deployment created"),
				Eventf(corev1.EventTypeNormal, dispatcherServiceCreated, "Dispatcher service created"),
				Eventf(corev1.EventTypeWarning, "InternalError", `endpoints "test-kc-dispatcher" not found`),
			},
		}, {
			Name: "Service does not exist, automatically created",
			Key:  kcKey,
			Objects: []runtime.Object{
				makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaFinalizer(finalizerName)),
			},
			WantErr: true,
			WantCreates: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaFinalizer(finalizerName),
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelDeploymentReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsNotReady("DispatcherEndpointsDoesNotExist", "Dispatcher Endpoints does not exist")),
			}},
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, dispatcherServiceAccountCreated, "Dispatcher service account created"),
				Eventf(corev1.EventTypeNormal, dispatcherRoleBindingCreated, "Dispatcher role binding created"),
				Eventf(corev1.EventTypeNormal, dispatcherRoleBindingCreated, "Dispatcher role binding created"),
				Eventf(corev1.EventTypeNormal, dispatcherServiceCreated, "Dispatcher service created"),
				Eventf(corev1.EventTypeWarning, "InternalError", `endpoints "test-kc-dispatcher" not found`),
			},
		}, {
			Name: "Endpoints does not exist",
			Key:  kcKey,
			Objects: []runtime.Object{
				makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaFinalizer(finalizerName)),
			},
			WantErr: true,
			WantCreates: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaFinalizer(finalizerName),
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelDeploymentReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsNotReady("DispatcherEndpointsDoesNotExist", "Dispatcher Endpoints does not exist"),
				),
			}},
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, dispatcherServiceAccountCreated, "Dispatcher service account created"),
				Eventf(corev1.EventTypeNormal, dispatcherRoleBindingCreated, "Dispatcher role binding created"),
				Eventf(corev1.EventTypeNormal, dispatcherRoleBindingCreated, "Dispatcher role binding created"),
				Eventf(corev1.EventTypeWarning, "InternalError", `endpoints "test-kc-dispatcher" not found`),
			},
		}, {
			Name: "Endpoints not ready",
			Key:  kcKey,
			Objects: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeEmptyEndpoints(),
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaFinalizer(finalizerName)),
			},
			WantErr: true,
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaFinalizer(finalizerName),
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelDeploymentReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsNotReady("DispatcherEndpointsNotReady", "There are no endpoints ready for Dispatcher service"),
				),
			}},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, "InternalError", `there are no endpoints ready for Dispatcher service test-kc-dispatcher`),
			},
		}, {
			Name: "Works, creates new channel",
			Key:  kcKey,
			Objects: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeReadyEndpoints(),
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaFinalizer(finalizerName)),
			},
			WantErr: false,
			WantCreates: []runtime.Object{
				makeChannelService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaFinalizer(finalizerName),
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelDeploymentReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsReady(),
					reconcilertesting.WithKafkaChannelChannelServiceReady(),
					reconcilertesting.WithKafkaChannelAddress(channelServiceAddress),
				),
			}},
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", `KafkaTopicChannel reconciled: "test-namespace/test-kc"`),
			},
		}, {
			Name: "Works, channel exists",
			Key:  kcKey,
			Objects: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeReadyEndpoints(),
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaChannelSubscribers(subscribers()),
					reconcilertesting.WithKafkaFinalizer(finalizerName)),
				makeChannelService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			},
			WantErr: false,
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaChannelSubscribers(subscribers()),
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaFinalizer(finalizerName),
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelDeploymentReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsReady(),
					reconcilertesting.WithKafkaChannelChannelServiceReady(),
					reconcilertesting.WithKafkaChannelAddress(channelServiceAddress),
				),
			}},
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", `KafkaTopicChannel reconciled: "test-namespace/test-kc"`),
			},
			WantPatches: []clientgotesting.PatchActionImpl{
				makePatch(testNS, kcName, twoSubscribersPatch),
			},
		}, {
			Name: "channel exists, not owned by us",
			Key:  kcKey,
			Objects: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeReadyEndpoints(),
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaFinalizer(finalizerName)),
				makeChannelServiceNotOwnedByUs(),
			},
			WantErr: true,
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaFinalizer(finalizerName),
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelDeploymentReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsReady(),
					reconcilertesting.WithKafkaChannelChannelServicetNotReady("ChannelServiceFailed", "Channel Service failed: kafkatopicchannel: test-namespace/test-kc does not own Service: \"test-kc-kn-channel\""),
				),
			}},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, "InternalError", `kafkatopicchannel: test-namespace/test-kc does not own Service: "test-kc-kn-channel"`),
			},
		}, {
			Name: "channel does not exist, fails to create",
			Key:  kcKey,
			Objects: []runtime.Object{
				makeServiceAccount(testNS, testDispatcherserviceAccount),
				makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
				makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
				makeReadyEndpoints(),
				reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithKafkaFinalizer(finalizerName)),
			},
			WantErr: true,
			WithReactors: []clientgotesting.ReactionFunc{
				InduceFailure("create", "Services"),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
					reconcilertesting.WithInitKafkaChannelConditions,
					reconcilertesting.WithKafkaFinalizer(finalizerName),
					reconcilertesting.WithKafkaChannelConfigReady(),
					reconcilertesting.WithKafkaChannelTopicReady(),
					reconcilertesting.WithKafkaChannelDeploymentReady(),
					reconcilertesting.WithKafkaChannelServiceReady(),
					reconcilertesting.WithKafkaChannelEndpointsReady(),
					reconcilertesting.WithKafkaChannelChannelServicetNotReady("ChannelServiceFailed", "Channel Service failed: inducing failure for create services"),
				),
			}},
			WantCreates: []runtime.Object{
				makeChannelService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, "InternalError", "inducing failure for create services"),
			},
			// 	// TODO add UTs for topic creation and deletion.
		},
	}

	table.Test(t, reconcilertesting.MakeFactory(func(ctx context.Context, listers *reconcilertesting.Listers, cmw configmap.Watcher) controller.Reconciler {

		r := &Reconciler{
			systemNamespace:          testNS,
			dispatcherImage:          testDispatcherImage,
			dispatcherServiceAccount: testDispatcherserviceAccount,
			kafkachannelLister:       listers.GetKafkaChannelLister(),
			// TODO fix
			kafkachannelInformer: nil,
			deploymentLister:     listers.GetDeploymentLister(),
			serviceLister:        listers.GetServiceLister(),
			endpointsLister:      listers.GetEndpointsLister(),
			serviceAccountLister: listers.GetServiceAccountLister(),
			roleBindingLister:    listers.GetRoleBindingLister(),
			kafkaClientSet:       fakekafkaclient.Get(ctx),
			KubeClientSet:        kubeclient.Get(ctx),
			EventingClientSet:    eventingClient.Get(ctx),
			statusManager: &fakeStatusManager{
				FakeIsReady: func(ctx context.Context, ch v1alpha1.KafkaTopicChannel,
					sub eventingduckv1.SubscriberSpec) (bool, error) {
					return true, nil
				},
			},
		}
		return kafkachannel.NewReconciler(ctx, logging.FromContext(ctx), r.kafkaClientSet, listers.GetKafkaChannelLister(), controller.GetEventRecorder(ctx), r)
	}, zap.L()))
}

func TestTopicExists(t *testing.T) {
	kcKey := testNS + "/" + kcName
	row := TableRow{
		Name: "Works, topic already exists",
		Key:  kcKey,
		Objects: []runtime.Object{
			makeServiceAccount(testNS, testDispatcherserviceAccount),
			makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeReadyDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			makeReadyEndpoints(),
			reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithKafkaFinalizer(finalizerName)),
		},
		WantErr: false,
		WantCreates: []runtime.Object{
			makeChannelService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithInitKafkaChannelConditions,
				reconcilertesting.WithKafkaFinalizer(finalizerName),
				reconcilertesting.WithKafkaChannelConfigReady(),
				reconcilertesting.WithKafkaChannelTopicReady(),
				reconcilertesting.WithKafkaChannelDeploymentReady(),
				reconcilertesting.WithKafkaChannelServiceReady(),
				reconcilertesting.WithKafkaChannelEndpointsReady(),
				reconcilertesting.WithKafkaChannelChannelServiceReady(),
				reconcilertesting.WithKafkaChannelAddress(channelServiceAddress),
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", `KafkaTopicChannel reconciled: "test-namespace/test-kc"`),
		},
	}

	row.Test(t, reconcilertesting.MakeFactory(func(ctx context.Context, listers *reconcilertesting.Listers, cmw configmap.Watcher) controller.Reconciler {

		r := &Reconciler{
			systemNamespace:          testNS,
			dispatcherImage:          testDispatcherImage,
			dispatcherServiceAccount: testDispatcherserviceAccount,
			kafkachannelLister:       listers.GetKafkaChannelLister(),
			// TODO fix
			kafkachannelInformer: nil,
			deploymentLister:     listers.GetDeploymentLister(),
			serviceLister:        listers.GetServiceLister(),
			endpointsLister:      listers.GetEndpointsLister(),
			serviceAccountLister: listers.GetServiceAccountLister(),
			roleBindingLister:    listers.GetRoleBindingLister(),
			kafkaClientSet:       fakekafkaclient.Get(ctx),
			KubeClientSet:        kubeclient.Get(ctx),
			EventingClientSet:    eventingClient.Get(ctx),
			statusManager: &fakeStatusManager{
				FakeIsReady: func(ctx context.Context, channel v1alpha1.KafkaTopicChannel,
					spec eventingduckv1.SubscriberSpec) (bool, error) {
					return true, nil
				},
			},
		}
		return kafkachannel.NewReconciler(ctx, logging.FromContext(ctx), r.kafkaClientSet, listers.GetKafkaChannelLister(), controller.GetEventRecorder(ctx), r)
	}, zap.L()))
}

func TestDeploymentUpdatedOnImageChange(t *testing.T) {
	kcKey := testNS + "/" + kcName
	row := TableRow{
		Name: "Works, topic already exists",
		Key:  kcKey,
		Objects: []runtime.Object{
			makeServiceAccount(testNS, testDispatcherserviceAccount),
			makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeDeploymentWithImageAndReplicas("differentimage", 1),
			makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			makeReadyEndpoints(),
			reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithKafkaFinalizer(finalizerName)),
		},
		WantErr: false,
		WantCreates: []runtime.Object{
			makeChannelService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: makeDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
		}},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithInitKafkaChannelConditions,
				reconcilertesting.WithKafkaFinalizer(finalizerName),
				reconcilertesting.WithKafkaChannelConfigReady(),
				reconcilertesting.WithKafkaChannelTopicReady(),
				//				reconcilekafkatesting.WithKafkaChannelDeploymentReady(),
				reconcilertesting.WithKafkaChannelServiceReady(),
				reconcilertesting.WithKafkaChannelEndpointsReady(),
				reconcilertesting.WithKafkaChannelChannelServiceReady(),
				reconcilertesting.WithKafkaChannelAddress(channelServiceAddress),
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeNormal, dispatcherDeploymentUpdated, "Dispatcher deployment updated"),
			Eventf(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", `KafkaTopicChannel reconciled: "test-namespace/test-kc"`),
		},
	}

	row.Test(t, reconcilertesting.MakeFactory(func(ctx context.Context, listers *reconcilertesting.Listers, cmw configmap.Watcher) controller.Reconciler {

		r := &Reconciler{
			systemNamespace:          testNS,
			dispatcherImage:          testDispatcherImage,
			dispatcherServiceAccount: testDispatcherserviceAccount,
			kafkachannelLister:       listers.GetKafkaChannelLister(),
			// TODO fix
			kafkachannelInformer: nil,
			deploymentLister:     listers.GetDeploymentLister(),
			serviceLister:        listers.GetServiceLister(),
			endpointsLister:      listers.GetEndpointsLister(),
			serviceAccountLister: listers.GetServiceAccountLister(),
			roleBindingLister:    listers.GetRoleBindingLister(),
			kafkaClientSet:       fakekafkaclient.Get(ctx),
			KubeClientSet:        kubeclient.Get(ctx),
			EventingClientSet:    eventingClient.Get(ctx),
			statusManager: &fakeStatusManager{
				FakeIsReady: func(ctx context.Context, channel v1alpha1.KafkaTopicChannel,
					spec eventingduckv1.SubscriberSpec) (bool, error) {
					return true, nil
				},
			},
		}
		return kafkachannel.NewReconciler(ctx, logging.FromContext(ctx), r.kafkaClientSet, listers.GetKafkaChannelLister(), controller.GetEventRecorder(ctx), r)
	}, zap.L()))
}

func TestDeploymentZeroReplicas(t *testing.T) {
	kcKey := testNS + "/" + kcName
	row := TableRow{
		Name: "Works, topic already exists",
		Key:  kcKey,
		Objects: []runtime.Object{
			makeServiceAccount(testNS, testDispatcherserviceAccount),
			makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeDeploymentWithImageAndReplicas(testDispatcherImage, 0),
			makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			makeReadyEndpoints(),
			reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithKafkaFinalizer(finalizerName)),
		},
		WantErr: false,
		WantCreates: []runtime.Object{
			makeChannelService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: makeDeployment(reconcilertesting.NewKafkaChannel(kcName, testNS)),
		}},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithInitKafkaChannelConditions,
				reconcilertesting.WithKafkaFinalizer(finalizerName),
				reconcilertesting.WithKafkaChannelConfigReady(),
				reconcilertesting.WithKafkaChannelTopicReady(),
				//				reconcilekafkatesting.WithKafkaChannelDeploymentReady(),
				reconcilertesting.WithKafkaChannelServiceReady(),
				reconcilertesting.WithKafkaChannelEndpointsReady(),
				reconcilertesting.WithKafkaChannelChannelServiceReady(),
				reconcilertesting.WithKafkaChannelAddress(channelServiceAddress),
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeNormal, dispatcherDeploymentUpdated, "Dispatcher deployment updated"),
			Eventf(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", `KafkaTopicChannel reconciled: "test-namespace/test-kc"`),
		},
	}

	row.Test(t, reconcilertesting.MakeFactory(func(ctx context.Context, listers *reconcilertesting.Listers, cmw configmap.Watcher) controller.Reconciler {

		r := &Reconciler{
			systemNamespace:          testNS,
			dispatcherImage:          testDispatcherImage,
			dispatcherServiceAccount: testDispatcherserviceAccount,
			kafkachannelLister:       listers.GetKafkaChannelLister(),
			// TODO fix
			kafkachannelInformer: nil,
			deploymentLister:     listers.GetDeploymentLister(),
			serviceLister:        listers.GetServiceLister(),
			endpointsLister:      listers.GetEndpointsLister(),
			serviceAccountLister: listers.GetServiceAccountLister(),
			roleBindingLister:    listers.GetRoleBindingLister(),
			kafkaClientSet:       fakekafkaclient.Get(ctx),
			KubeClientSet:        kubeclient.Get(ctx),
			EventingClientSet:    eventingClient.Get(ctx),
			statusManager: &fakeStatusManager{
				FakeIsReady: func(ctx context.Context, channel v1alpha1.KafkaTopicChannel,
					spec eventingduckv1.SubscriberSpec) (bool, error) {
					return true, nil
				},
			},
		}
		return kafkachannel.NewReconciler(ctx, logging.FromContext(ctx), r.kafkaClientSet, listers.GetKafkaChannelLister(), controller.GetEventRecorder(ctx), r)
	}, zap.L()))
}

func TestDeploymentMoreThanOneReplicas(t *testing.T) {
	kcKey := testNS + "/" + kcName
	row := TableRow{
		Name: "Works, topic already exists",
		Key:  kcKey,
		Objects: []runtime.Object{
			makeServiceAccount(testNS, testDispatcherserviceAccount),
			makeRoleBinding(testNS, testDispatcherCR, testDispatcherCR, makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeRoleBinding(testNS, testEventConfigReadCRName, "eventing-config-reader", makeServiceAccount(testNS, testDispatcherserviceAccount)),
			makeDeploymentWithImageAndReplicas(testDispatcherImage, 3),
			makeService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
			makeReadyEndpoints(),
			reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithKafkaFinalizer(finalizerName)),
		},
		WantErr: false,
		WantCreates: []runtime.Object{
			makeChannelService(reconcilertesting.NewKafkaChannel(kcName, testNS)),
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: reconcilertesting.NewKafkaChannel(kcName, testNS,
				reconcilertesting.WithInitKafkaChannelConditions,
				reconcilertesting.WithKafkaFinalizer(finalizerName),
				reconcilertesting.WithKafkaChannelConfigReady(),
				reconcilertesting.WithKafkaChannelTopicReady(),
				//				reconcilekafkatesting.WithKafkaChannelDeploymentReady(),
				reconcilertesting.WithKafkaChannelServiceReady(),
				reconcilertesting.WithKafkaChannelEndpointsReady(),
				reconcilertesting.WithKafkaChannelChannelServiceReady(),
				reconcilertesting.WithKafkaChannelAddress(channelServiceAddress),
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", `KafkaTopicChannel reconciled: "test-namespace/test-kc"`),
		},
	}

	row.Test(t, reconcilertesting.MakeFactory(func(ctx context.Context, listers *reconcilertesting.Listers, cmw configmap.Watcher) controller.Reconciler {

		r := &Reconciler{
			systemNamespace:          testNS,
			dispatcherImage:          testDispatcherImage,
			dispatcherServiceAccount: testDispatcherserviceAccount,
			kafkachannelLister:       listers.GetKafkaChannelLister(),
			// TODO fix
			kafkachannelInformer: nil,
			deploymentLister:     listers.GetDeploymentLister(),
			serviceLister:        listers.GetServiceLister(),
			endpointsLister:      listers.GetEndpointsLister(),
			serviceAccountLister: listers.GetServiceAccountLister(),
			roleBindingLister:    listers.GetRoleBindingLister(),
			kafkaClientSet:       fakekafkaclient.Get(ctx),
			KubeClientSet:        kubeclient.Get(ctx),
			EventingClientSet:    eventingClient.Get(ctx),
			statusManager: &fakeStatusManager{
				FakeIsReady: func(ctx context.Context, channel v1alpha1.KafkaTopicChannel,
					spec eventingduckv1.SubscriberSpec) (bool, error) {
					return true, nil
				},
			},
		}
		return kafkachannel.NewReconciler(ctx, logging.FromContext(ctx), r.kafkaClientSet, listers.GetKafkaChannelLister(), controller.GetEventRecorder(ctx), r)
	}, zap.L()))
}

func makeDeploymentWithParams(image string, replicas int32, nc *v1alpha1.KafkaTopicChannel) *appsv1.Deployment {
	return resources.MakeDispatcher(resources.DispatcherArgs{
		DispatcherNamespace: testNS,
		Image:               image,
		Replicas:            replicas,
		ServiceAccount:      testDispatcherserviceAccount,
		Channel:             nc,
		DispatcherScope:     "namespace",
	})
}

func makeDeploymentWithImageAndReplicas(image string, replicas int32) *appsv1.Deployment {
	return makeDeploymentWithParams(image, replicas, reconcilertesting.NewKafkaChannel(kcName, testNS))
}

func makeDeploymentWithConfigMapHash(nc *v1alpha1.KafkaTopicChannel) *appsv1.Deployment {
	return makeDeploymentWithParams(testDispatcherImage, 1, nc)
}

func makeDeployment(nc *v1alpha1.KafkaTopicChannel) *appsv1.Deployment {
	return makeDeploymentWithImageAndReplicas(testDispatcherImage, 1)
}

func makeServiceAccount(ns, name string) *corev1.ServiceAccount {
	return resources.MakeServiceAccount(ns, name)
}

func makeRoleBinding(ns, name, role string, sa *corev1.ServiceAccount) *rbacv1.RoleBinding {
	return resources.MakeRoleBinding(ns, name, sa, role)
}

func makeReadyDeployment(nc *v1alpha1.KafkaTopicChannel) *appsv1.Deployment {
	d := makeDeployment(nc)
	d.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue}}
	return d
}

func makeService(nc *v1alpha1.KafkaTopicChannel) *corev1.Service {
	return resources.MakeDispatcherService(testNS, nc)
}

func makeChannelService(nc *v1alpha1.KafkaTopicChannel) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNS,
			Name:      fmt.Sprintf("%s-kn-channel", kcName),
			Labels: map[string]string{
				resources.MessagingRoleLabel: resources.MessagingRole,
			},
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(nc),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: network.GetServiceHostname(testDispatcherName, testNS),
		},
	}
}

func makeChannelServiceNotOwnedByUs() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNS,
			Name:      fmt.Sprintf("%s-kn-channel", kcName),
			Labels: map[string]string{
				resources.MessagingRoleLabel: resources.MessagingRole,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: network.GetServiceHostname(testDispatcherName, testNS),
		},
	}
}

func makeEmptyEndpoints() *corev1.Endpoints {
	return &corev1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Endpoints",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNS,
			Name:      testDispatcherName,
		},
	}
}

func makeReadyEndpoints() *corev1.Endpoints {
	e := makeEmptyEndpoints()
	e.Subsets = []corev1.EndpointSubset{{Addresses: []corev1.EndpointAddress{{IP: "1.1.1.1"}}}}
	return e
}

func patchFinalizers(namespace, name string) clientgotesting.PatchActionImpl {
	action := clientgotesting.PatchActionImpl{}
	action.Name = name
	action.Namespace = namespace
	patch := `{"metadata":{"finalizers":["` + finalizerName + `"],"resourceVersion":""}}`
	action.Patch = []byte(patch)
	return action
}

func subscribers() []eventingduckv1.SubscriberSpec {

	return []eventingduckv1.SubscriberSpec{{
		UID:           sub1UID,
		Generation:    1,
		SubscriberURI: apis.HTTP("call1"),
		ReplyURI:      apis.HTTP("sink2"),
	}, {
		UID:           sub2UID,
		Generation:    2,
		SubscriberURI: apis.HTTP("call2"),
		ReplyURI:      apis.HTTP("sink2"),
	}}
}

type fakeStatusManager struct {
	FakeIsReady func(context.Context, v1alpha1.KafkaTopicChannel, eventingduckv1.SubscriberSpec) (bool, error)
}

func (m *fakeStatusManager) IsReady(ctx context.Context, ch v1alpha1.KafkaTopicChannel, sub eventingduckv1.SubscriberSpec) (bool, error) {
	return m.FakeIsReady(ctx, ch, sub)
}

func (m *fakeStatusManager) CancelProbing(sub eventingduckv1.SubscriberSpec) {
	//do nothing
}

func (m *fakeStatusManager) CancelPodProbing(pod corev1.Pod) {
	//do nothing
}

func makePatch(namespace, name, patch string) clientgotesting.PatchActionImpl {
	return clientgotesting.PatchActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Namespace: namespace,
		},
		Name:  name,
		Patch: []byte(patch),
	}
}

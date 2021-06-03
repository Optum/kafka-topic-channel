package controller

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"knative.dev/eventing-kafka/pkg/common/constants"
	"knative.dev/eventing/pkg/channel/fanout"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/configmap"
	configmapinformer "knative.dev/pkg/configmap/informer"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/system"
	"knative.dev/pkg/tracing"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	"github.com/optum/kafka-topic-channel/pkg/channel/kt/dispatcher"
	"github.com/optum/kafka-topic-channel/pkg/channel/kt/utils"
	kafkaclientset "github.com/optum/kafka-topic-channel/pkg/client/clientset/versioned"
	kafkaScheme "github.com/optum/kafka-topic-channel/pkg/client/clientset/versioned/scheme"
	kafkaclientsetinjection "github.com/optum/kafka-topic-channel/pkg/client/injection/client"
	"github.com/optum/kafka-topic-channel/pkg/client/injection/informers/messaging/v1alpha1/kafkatopicchannel"
	kafkachannelreconciler "github.com/optum/kafka-topic-channel/pkg/client/injection/reconciler/messaging/v1alpha1/kafkatopicchannel"
	listers "github.com/optum/kafka-topic-channel/pkg/client/listers/messaging/v1alpha1"
)

func init() {
	// Add run types to the default Kubernetes Scheme so Events can be
	// logged for run types.
	_ = kafkaScheme.AddToScheme(scheme.Scheme)
}

// Reconciler reconciles Kafka Channels.
type Reconciler struct {
	kafkaDispatcher *dispatcher.KafkaDispatcher

	kafkaClientSet       kafkaclientset.Interface
	kafkachannelLister   listers.KafkaTopicChannelLister
	kafkachannelInformer cache.SharedIndexInformer
	impl                 *controller.Impl
}

var _ kafkachannelreconciler.Interface = (*Reconciler)(nil)
var channelName = os.Getenv("CHANNEL_NAME")

// NewController initializes the controller and is called by the generated code.
// Registers event handlers to enqueue events.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	dispatcherName := fmt.Sprintf("%s-dispatcher", channelName)
	logger := logging.FromContext(ctx)

	err := tracing.SetupDynamicPublishing(logger, cmw.(*configmapinformer.InformedWatcher), dispatcherName, "config-tracing")
	if err != nil {
		logger.Fatalw("unable to setup tracing", zap.Error(err))
	}

	// Configure connection arguments - to be done exactly once per process
	kncloudevents.ConfigureConnectionArgs(&kncloudevents.ConnectionArgs{
		MaxIdleConns:        int(constants.DefaultMaxIdleConns),
		MaxIdleConnsPerHost: int(constants.DefaultMaxIdleConnsPerHost),
	})
	brokers, cfg, err := utils.NewConfigFromEnv(ctx)
	if err != nil {
		logger.Fatalw("unable to create sarama config", zap.Error(err))
	}
	kafkaChannelInformer := kafkatopicchannel.Get(ctx)
	args := &dispatcher.KafkaDispatcherArgs{
		ClientID:    dispatcherName,
		Brokers:     brokers,
		KafkaConfig: cfg,
		TopicFunc:   utils.TopicName,
	}
	kafkaDispatcher, err := dispatcher.NewDispatcher(ctx, args)
	if err != nil {
		logger.Fatalw("Unable to create kafka dispatcher", zap.Error(err))
	}
	logger.Info("Starting the Kafka dispatcher")

	r := &Reconciler{
		kafkaDispatcher:      kafkaDispatcher,
		kafkaClientSet:       kafkaclientsetinjection.Get(ctx),
		kafkachannelLister:   kafkaChannelInformer.Lister(),
		kafkachannelInformer: kafkaChannelInformer.Informer(),
	}
	r.impl = kafkachannelreconciler.NewImpl(ctx, r, func(impl *controller.Impl) controller.Options {
		return controller.Options{SkipStatusUpdates: true}
	})

	logger.Info("Setting up event handlers")

	// Watch for kafka channels.
	kafkaChannelInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: filterWithAnnotation(),
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    r.impl.Enqueue,
				UpdateFunc: controller.PassNew(r.impl.Enqueue),
				DeleteFunc: func(obj interface{}) {
					// TODO when finalize kind is fixed, we'll need to handle that error properly
					if err := r.CleanupChannel(obj.(*v1alpha1.KafkaTopicChannel)); err != nil {
						logger.Warnw("Unable to remove kafka channel", zap.Any("kafkachannel", obj), zap.Error(err))
					}
				},
			},
		})

	logger.Info("Starting dispatcher.")
	go func() {
		if err := kafkaDispatcher.Start(ctx); err != nil {
			logger.Errorw("Cannot start dispatcher", zap.Error(err))
		}
	}()

	return r.impl
}

func filterWithAnnotation() func(obj interface{}) bool {
	return NameAndNamespaceFilterFunc(channelName, system.Namespace())
}

func NameAndNamespaceFilterFunc(name, namespace string) func(interface{}) bool {
	return func(obj interface{}) bool {
		if mo, ok := obj.(metav1.Object); ok {
			return mo.GetNamespace() == namespace && mo.GetName() == name
		}
		return false
	}
}

func (r *Reconciler) ReconcileKind(ctx context.Context, kc *v1alpha1.KafkaTopicChannel) pkgreconciler.Event {
	logging.FromContext(ctx).Debugw("ReconcileKind for channel", zap.String("channel", kc.Name))
	return r.syncChannel(ctx, kc)
}

func (r *Reconciler) ObserveKind(ctx context.Context, kc *v1alpha1.KafkaTopicChannel) pkgreconciler.Event {
	logging.FromContext(ctx).Debugw("ObserveKind for channel", zap.String("channel", kc.Name))
	return r.syncChannel(ctx, kc)
}

func (r *Reconciler) syncChannel(ctx context.Context, kc *v1alpha1.KafkaTopicChannel) pkgreconciler.Event {
	if !kc.Status.IsReady() {
		logging.FromContext(ctx).Debugw("KafkaChannel still not ready, short-circuiting the reconciler", zap.String("channel", kc.Name))
		return nil
	}
	config := r.newConfigFromKafkaChannel(kc)

	// Update receiver side
	if err := r.kafkaDispatcher.RegisterChannelHost(config); err != nil {
		logging.FromContext(ctx).Error("Error updating host to channel map in dispatcher")
		return err
	}

	// Update dispatcher side
	err := r.kafkaDispatcher.ReconcileConsumers(config)
	if err != nil {
		logging.FromContext(ctx).Errorw("Some kafka subscriptions failed to subscribe", zap.Error(err))
		return fmt.Errorf("some kafka subscriptions failed to subscribe: %v", err)
	}
	return nil
}

func (r *Reconciler) CleanupChannel(kc *v1alpha1.KafkaTopicChannel) pkgreconciler.Event {
	return r.kafkaDispatcher.CleanupChannel(kc.Name, kc.Namespace, kc.Status.Address.URL.Host)
}

// newConfigFromKafkaChannel creates a new Config from the list of kafka channels.
func (r *Reconciler) newConfigFromKafkaChannel(c *v1alpha1.KafkaTopicChannel) *dispatcher.ChannelConfig {
	channelConfig := dispatcher.ChannelConfig{
		Namespace: c.Namespace,
		Name:      c.Name,
		HostName:  c.Status.Address.URL.Host,
	}
	if c.Spec.SubscribableSpec.Subscribers != nil {
		newSubs := make([]dispatcher.Subscription, 0, len(c.Spec.SubscribableSpec.Subscribers))
		for _, source := range c.Spec.SubscribableSpec.Subscribers {
			innerSub, _ := fanout.SubscriberSpecToFanoutConfig(source)

			newSubs = append(newSubs, dispatcher.Subscription{
				Subscription: *innerSub,
				UID:          source.UID,
			})
		}
		channelConfig.Subscriptions = newSubs
	}

	return &channelConfig
}

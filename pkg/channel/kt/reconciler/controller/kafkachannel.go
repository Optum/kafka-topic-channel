package controller

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/pointer"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	"github.com/optum/kafka-topic-channel/pkg/channel/kt/reconciler/controller/resources"
	"github.com/optum/kafka-topic-channel/pkg/channel/kt/status"
	"github.com/optum/kafka-topic-channel/pkg/channel/kt/utils"
	kafkaclientset "github.com/optum/kafka-topic-channel/pkg/client/clientset/versioned"
	kafkaScheme "github.com/optum/kafka-topic-channel/pkg/client/clientset/versioned/scheme"
	kafkaChannelReconciler "github.com/optum/kafka-topic-channel/pkg/client/injection/reconciler/messaging/v1alpha1/kafkatopicchannel"
	listers "github.com/optum/kafka-topic-channel/pkg/client/listers/messaging/v1alpha1"
	v1 "knative.dev/eventing/pkg/apis/duck/v1"
	eventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	pkgreconciler "knative.dev/pkg/reconciler"
)

const (
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "kt-channel-controller"
	dispatcherCR        = "kt-channel-dispatcher"
	dispatcherSA        = "kt-channel-dispatcher"

	// Name of the corev1.Events emitted from the reconciliation process.
	dispatcherDeploymentCreated     = "DispatcherDeploymentCreated"
	dispatcherDeploymentUpdated     = "DispatcherDeploymentUpdated"
	dispatcherDeploymentFailed      = "DispatcherDeploymentFailed"
	dispatcherServiceCreated        = "DispatcherServiceCreated"
	dispatcherServiceFailed         = "DispatcherServiceFailed"
	dispatcherServiceAccountCreated = "DispatcherServiceAccountCreated"
	dispatcherRoleBindingCreated    = "DispatcherRoleBindingCreated"
)

var (
	scopeNamespace = "namespace"
)

func newReconciledNormal(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "KafkaTopicChannelReconciled", "KafkaTopicChannel reconciled: \"%s/%s\"", namespace, name)
}

func newDeploymentWarn(err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DispatcherDeploymentFailed", "Reconciling dispatcher Deployment failed with: %s", err)
}

func newDispatcherServiceWarn(err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DispatcherServiceFailed", "Reconciling dispatcher Service failed with: %s", err)
}

func newServiceAccountWarn(err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DispatcherServiceAccountFailed", "Reconciling dispatcher ServiceAccount failed: %s", err)
}

func newRoleBindingWarn(err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DispatcherRoleBindingFailed", "Reconciling dispatcher RoleBinding failed: %s", err)
}

func init() {
	// Add run types to the default Kubernetes Scheme so Events can be
	// logged for run types.
	_ = kafkaScheme.AddToScheme(scheme.Scheme)
}

// Reconciler reconciles Kafka Channels.
type Reconciler struct {
	KubeClientSet kubernetes.Interface

	EventingClientSet eventingclientset.Interface

	systemNamespace          string
	dispatcherImage          string
	dispatcherServiceAccount string

	kafkaClientSet       kafkaclientset.Interface
	kafkachannelLister   listers.KafkaTopicChannelLister
	kafkachannelInformer cache.SharedIndexInformer
	deploymentLister     appsv1listers.DeploymentLister
	serviceLister        corev1listers.ServiceLister
	endpointsLister      corev1listers.EndpointsLister
	serviceAccountLister corev1listers.ServiceAccountLister
	roleBindingLister    rbacv1listers.RoleBindingLister
	statusManager        status.Manager
}

type envConfig struct {
	Image string `envconfig:"DISPATCHER_IMAGE" required:"true"`
}

// Check that our Reconciler implements kafka's injection Interface
var _ kafkaChannelReconciler.Interface = (*Reconciler)(nil)
var _ kafkaChannelReconciler.Finalizer = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, kc *v1alpha1.KafkaTopicChannel) pkgreconciler.Event {
	dispatcherName := fmt.Sprintf("%s-dispatcher", kc.Name)
	kc.Status.InitializeConditions()
	logger := logging.FromContext(ctx)
	// Verify channel is valid.
	kc.SetDefaults(ctx)
	if err := kc.Validate(ctx); err != nil {
		logger.Errorw("Invalid kafka channel", zap.String("channel", kc.Name), zap.Error(err))
		return err
	}

	kc.Status.MarkConfigTrue()

	kc.Status.MarkTopicTrue()

	dispatcherNamespace := kc.Namespace

	// Make sure the dispatcher deployment exists and propagate the status to the Channel
	_, err := r.reconcileDispatcher(ctx, dispatcherNamespace, kc)
	if err != nil {
		return err
	}

	// Make sure the dispatcher service exists and propagate the status to the Channel in case it does not exist.
	// We don't do anything with the service because it's status contains nothing useful, so just do
	// an existence check. Then below we check the endpoints targeting it.
	_, err = r.reconcileDispatcherService(ctx, dispatcherNamespace, kc)
	if err != nil {
		return err
	}

	// Get the Dispatcher Service Endpoints and propagate the status to the Channel
	// endpoints has the same name as the service, so not a bug.
	e, err := r.endpointsLister.Endpoints(dispatcherNamespace).Get(dispatcherName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			kc.Status.MarkEndpointsFailed("DispatcherEndpointsDoesNotExist", "Dispatcher Endpoints does not exist")
		} else {
			logger.Errorw("Unable to get the dispatcher endpoints", zap.Error(err))
			kc.Status.MarkEndpointsFailed("DispatcherEndpointsGetFailed", "Failed to get dispatcher endpoints")
		}
		return err
	}

	if len(e.Subsets) == 0 {
		logger.Errorw("No endpoints found for Dispatcher service", zap.Error(err))
		kc.Status.MarkEndpointsFailed("DispatcherEndpointsNotReady", "There are no endpoints ready for Dispatcher service")
		return fmt.Errorf("there are no endpoints ready for Dispatcher service %s", dispatcherName)
	}
	kc.Status.MarkEndpointsTrue()

	// Reconcile the k8s service representing the actual Channel. It points to the Dispatcher service via ExternalName
	svc, err := r.reconcileChannelService(ctx, dispatcherNamespace, kc)
	if err != nil {
		return err
	}
	kc.Status.MarkChannelServiceTrue()
	kc.Status.SetAddress(&apis.URL{
		Scheme: "http",
		Host:   network.GetServiceHostname(svc.Name, svc.Namespace),
	})
	err = r.reconcileSubscribers(ctx, kc)
	if err != nil {
		return fmt.Errorf("error reconciling subscribers %v", err)
	}

	// Ok, so now the Dispatcher Deployment & Service have been created, we're golden since the
	// dispatcher watches the Channel and where it needs to dispatch events to.
	return newReconciledNormal(kc.Namespace, kc.Name)
}

func (r *Reconciler) reconcileSubscribers(ctx context.Context, ch *v1alpha1.KafkaTopicChannel) error {
	after := ch.DeepCopy()
	after.Status.Subscribers = make([]v1.SubscriberStatus, 0)
	for _, s := range ch.Spec.Subscribers {
		if r, _ := r.statusManager.IsReady(ctx, *ch, s); r {
			logging.FromContext(ctx).Debugw("marking subscription", zap.Any("subscription", s))
			after.Status.Subscribers = append(after.Status.Subscribers, v1.SubscriberStatus{
				UID:                s.UID,
				ObservedGeneration: s.Generation,
				Ready:              corev1.ConditionTrue,
			})
		}
	}

	jsonPatch, err := duck.CreatePatch(ch, after)
	if err != nil {
		return fmt.Errorf("creating JSON patch: %w", err)
	}
	// If there is nothing to patch, we are good, just return.
	// Empty patch is [], hence we check for that.
	if len(jsonPatch) == 0 {
		return nil
	}
	patch, err := jsonPatch.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshaling JSON patch: %w", err)
	}
	patched, err := r.kafkaClientSet.MessagingV1alpha1().KafkaTopicChannels(ch.Namespace).Patch(ctx, ch.Name, types.JSONPatchType, patch, metav1.PatchOptions{}, "status")

	if err != nil {
		return fmt.Errorf("Failed patching: %w", err)
	}
	logging.FromContext(ctx).Debugw("Patched resource", zap.Any("patch", patch), zap.Any("patched", patched))
	return nil
}

func (r *Reconciler) reconcileDispatcher(ctx context.Context, dispatcherNamespace string, kc *v1alpha1.KafkaTopicChannel) (*appsv1.Deployment, error) {
	// Configure RBAC in namespace to access the configmaps
	sa, err := r.reconcileServiceAccount(ctx, dispatcherNamespace, kc)
	if err != nil {
		return nil, err
	}

	_, err = r.reconcileRoleBinding(ctx, dispatcherCR, dispatcherNamespace, kc, dispatcherCR, sa)
	if err != nil {
		return nil, err
	}

	// Reconcile the RoleBinding allowing read access to the shared configmaps.
	// Note this RoleBinding is created in the system namespace and points to a
	// subject in the dispatcher's namespace.
	// TODO: might change when ConfigMapPropagation lands
	roleBindingName := fmt.Sprintf("%s-%s", dispatcherCR, dispatcherNamespace)
	_, err = r.reconcileRoleBinding(ctx, roleBindingName, r.systemNamespace, kc, "eventing-config-reader", sa)
	if err != nil {
		return nil, err
	}
	args := resources.DispatcherArgs{
		DispatcherScope:     scopeNamespace,
		DispatcherNamespace: dispatcherNamespace,
		Image:               r.dispatcherImage,
		Replicas:            1,
		ServiceAccount:      r.dispatcherServiceAccount,
		Channel:             kc,
	}

	expected := resources.MakeDispatcher(args)
	d, err := r.deploymentLister.Deployments(dispatcherNamespace).Get(fmt.Sprintf("%s-dispatcher", kc.Name))
	if err != nil {
		if apierrs.IsNotFound(err) {
			d, err := r.KubeClientSet.AppsV1().Deployments(dispatcherNamespace).Create(ctx, expected, metav1.CreateOptions{})
			if err == nil {
				controller.GetEventRecorder(ctx).Event(kc, corev1.EventTypeNormal, dispatcherDeploymentCreated, "Dispatcher deployment created")
				kc.Status.PropagateDispatcherStatus(&d.Status)
				return d, err
			} else {
				kc.Status.MarkDispatcherFailed(dispatcherDeploymentFailed, "Failed to create the dispatcher deployment: %v", err)
				return d, newDeploymentWarn(err)
			}
		}

		logging.FromContext(ctx).Errorw("Unable to get the dispatcher deployment", zap.Error(err))
		kc.Status.MarkDispatcherUnknown("DispatcherDeploymentFailed", "Failed to get dispatcher deployment: %v", err)
		return nil, err
	} else {
		existing := utils.FindContainer(d, resources.DispatcherContainerName)
		if existing == nil {
			logging.FromContext(ctx).Errorw("Container %s does not exist in existing dispatcher deployment. Updating the deployment", resources.DispatcherContainerName)
			d, err := r.KubeClientSet.AppsV1().Deployments(dispatcherNamespace).Update(ctx, expected, metav1.UpdateOptions{})
			if err == nil {
				controller.GetEventRecorder(ctx).Event(kc, corev1.EventTypeNormal, dispatcherDeploymentUpdated, "Dispatcher deployment updated")
				kc.Status.PropagateDispatcherStatus(&d.Status)
				return d, nil
			} else {
				kc.Status.MarkServiceFailed("DispatcherDeploymentUpdateFailed", "Failed to update the dispatcher deployment: %v", err)
			}
			return d, newDeploymentWarn(err)
		}

		expectedContainer := utils.FindContainer(expected, resources.DispatcherContainerName)
		if expectedContainer == nil {
			return nil, fmt.Errorf("container %s does not exist in expected dispatcher deployment. Cannot check if the deployment needs an update", resources.DispatcherContainerName)
		}

		needsUpdate := false

		if existing.Image != expectedContainer.Image {
			logging.FromContext(ctx).Infof("Dispatcher deployment image is not what we expect it to be, updating Deployment Got: %q Expect: %q", expected.Spec.Template.Spec.Containers[0].Image, d.Spec.Template.Spec.Containers[0].Image)
			existing.Image = expectedContainer.Image
			needsUpdate = true
		}

		if *d.Spec.Replicas == 0 {
			logging.FromContext(ctx).Infof("Dispatcher deployment has 0 replica. Scaling up deployment to 1 replica")
			d.Spec.Replicas = pointer.Int32Ptr(1)
			needsUpdate = true
		}

		if needsUpdate {
			d, err := r.KubeClientSet.AppsV1().Deployments(dispatcherNamespace).Update(ctx, d, metav1.UpdateOptions{})
			if err == nil {
				controller.GetEventRecorder(ctx).Event(kc, corev1.EventTypeNormal, dispatcherDeploymentUpdated, "Dispatcher deployment updated")
				kc.Status.PropagateDispatcherStatus(&d.Status)
				return d, nil
			} else {
				kc.Status.MarkServiceFailed("DispatcherDeploymentUpdateFailed", "Failed to update the dispatcher deployment: %v", err)
				return d, newDeploymentWarn(err)
			}
		}
	}

	kc.Status.PropagateDispatcherStatus(&d.Status)
	return d, nil
}

func (r *Reconciler) reconcileServiceAccount(ctx context.Context, dispatcherNamespace string, kc *v1alpha1.KafkaTopicChannel) (*corev1.ServiceAccount, error) {
	sa, err := r.serviceAccountLister.ServiceAccounts(dispatcherNamespace).Get(r.dispatcherServiceAccount)
	if err != nil {
		if apierrs.IsNotFound(err) {
			expected := resources.MakeServiceAccount(dispatcherNamespace, r.dispatcherServiceAccount)
			sa, err := r.KubeClientSet.CoreV1().ServiceAccounts(dispatcherNamespace).Create(ctx, expected, metav1.CreateOptions{})
			if err == nil {
				controller.GetEventRecorder(ctx).Event(kc, corev1.EventTypeNormal, dispatcherServiceAccountCreated, "Dispatcher service account created")
				return sa, nil
			} else {
				kc.Status.MarkDispatcherFailed("DispatcherDeploymentFailed", "Failed to create the dispatcher service account: %v", err)
				return sa, newServiceAccountWarn(err)
			}
		}

		kc.Status.MarkDispatcherUnknown("DispatcherServiceAccountFailed", "Failed to get dispatcher service account: %v", err)
		return nil, newServiceAccountWarn(err)
	}
	return sa, err
}

func (r *Reconciler) reconcileRoleBinding(ctx context.Context, name string, ns string, kc *v1alpha1.KafkaTopicChannel, clusterRoleName string, sa *corev1.ServiceAccount) (*rbacv1.RoleBinding, error) {
	rb, err := r.roleBindingLister.RoleBindings(ns).Get(name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			expected := resources.MakeRoleBinding(ns, name, sa, clusterRoleName)
			rb, err := r.KubeClientSet.RbacV1().RoleBindings(ns).Create(ctx, expected, metav1.CreateOptions{})
			if err == nil {
				controller.GetEventRecorder(ctx).Event(kc, corev1.EventTypeNormal, dispatcherRoleBindingCreated, "Dispatcher role binding created")
				return rb, nil
			} else {
				kc.Status.MarkDispatcherFailed("DispatcherDeploymentFailed", "Failed to create the dispatcher role binding: %v", err)
				return rb, newRoleBindingWarn(err)
			}
		}
		kc.Status.MarkDispatcherUnknown("DispatcherRoleBindingFailed", "Failed to get dispatcher role binding: %v", err)
		return nil, newRoleBindingWarn(err)
	}
	return rb, err
}

func (r *Reconciler) reconcileDispatcherService(ctx context.Context, dispatcherNamespace string, kc *v1alpha1.KafkaTopicChannel) (*corev1.Service, error) {
	svc, err := r.serviceLister.Services(dispatcherNamespace).Get(fmt.Sprintf("%s-dispatcher", kc.Name))
	if err != nil {
		if apierrs.IsNotFound(err) {
			expected := resources.MakeDispatcherService(dispatcherNamespace, kc)
			svc, err := r.KubeClientSet.CoreV1().Services(dispatcherNamespace).Create(ctx, expected, metav1.CreateOptions{})

			if err == nil {
				controller.GetEventRecorder(ctx).Event(kc, corev1.EventTypeNormal, dispatcherServiceCreated, "Dispatcher service created")
				kc.Status.MarkServiceTrue()
			} else {
				logging.FromContext(ctx).Errorw("Unable to create the dispatcher service", zap.Error(err))
				controller.GetEventRecorder(ctx).Eventf(kc, corev1.EventTypeWarning, dispatcherServiceFailed, "Failed to create the dispatcher service: %v", err)
				kc.Status.MarkServiceFailed("DispatcherServiceFailed", "Failed to create the dispatcher service: %v", err)
				return svc, err
			}

			return svc, err
		}

		kc.Status.MarkServiceUnknown("DispatcherServiceFailed", "Failed to get dispatcher service: %v", err)
		return nil, newDispatcherServiceWarn(err)
	}

	kc.Status.MarkServiceTrue()
	return svc, nil
}

func (r *Reconciler) reconcileChannelService(ctx context.Context, dispatcherNamespace string, channel *v1alpha1.KafkaTopicChannel) (*corev1.Service, error) {
	logger := logging.FromContext(ctx)
	// Get the  Service and propagate the status to the Channel in case it does not exist.
	// We don't do anything with the service because it's status contains nothing useful, so just do
	// an existence check. Then below we check the endpoints targeting it.
	// We may change this name later, so we have to ensure we use proper addressable when resolving these.
	expected, err := resources.MakeK8sService(channel, resources.ExternalService(dispatcherNamespace, fmt.Sprintf("%s-dispatcher", channel.Name)))
	if err != nil {
		logger.Errorw("failed to create the channel service object", zap.Error(err))
		channel.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service failed: %s", err))
		return nil, err
	}

	svc, err := r.serviceLister.Services(channel.Namespace).Get(resources.MakeChannelServiceName(channel.Name))
	if err != nil {
		if apierrs.IsNotFound(err) {
			svc, err = r.KubeClientSet.CoreV1().Services(channel.Namespace).Create(ctx, expected, metav1.CreateOptions{})
			if err != nil {
				logger.Errorw("failed to create the channel service object", zap.Error(err))
				channel.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service failed: %s", err))
				return nil, err
			}
			return svc, nil
		}
		logger.Errorw("Unable to get the channel service", zap.Error(err))
		return nil, err
	} else if !equality.Semantic.DeepEqual(svc.Spec, expected.Spec) {
		svc = svc.DeepCopy()
		svc.Spec = expected.Spec

		svc, err = r.KubeClientSet.CoreV1().Services(channel.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
		if err != nil {
			logger.Errorw("Failed to update the channel service", zap.Error(err))
			return nil, err
		}
	}
	// Check to make sure that the KafkaChannel owns this service and if not, complain.
	if !metav1.IsControlledBy(svc, channel) {
		err := fmt.Errorf("kafkatopicchannel: %s/%s does not own Service: %q", channel.Namespace, channel.Name, svc.Name)
		channel.Status.MarkChannelServiceFailed("ChannelServiceFailed", "Channel Service failed: %s", err)
		return nil, err
	}
	return svc, nil
}

func (r *Reconciler) FinalizeKind(ctx context.Context, kc *v1alpha1.KafkaTopicChannel) pkgreconciler.Event {
	// Do not attempt retrying creating the client because it might be a permanent error
	// in which case the finalizer will never get removed.
	logger := logging.FromContext(ctx)
	channel := fmt.Sprintf("%s/%s", kc.GetNamespace(), kc.GetName())
	logger.Debugw("FinalizeKind", zap.String("channel", channel))

	for _, s := range kc.Spec.Subscribers {
		logger.Debugw("Canceling probing", zap.String("channel", channel), zap.Any("subscription", s))
		r.statusManager.CancelProbing(s)
	}
	return newReconciledNormal(kc.Namespace, kc.Name) //ok to remove finalizer
}

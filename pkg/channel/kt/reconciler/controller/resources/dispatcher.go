package resources

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/system"
)

const (
	DispatcherContainerName = "dispatcher"
)

var (
	dispatcherName   string
	dispatcherLabels = map[string]string{
		"messaging.optum.dev/channel": "kafka-topic-channel",
		"messaging.optum.dev/role":    "dispatcher",
	}
)

type DispatcherArgs struct {
	DispatcherScope     string
	DispatcherNamespace string
	Image               string
	Replicas            int32
	ServiceAccount      string
	ConfigMapHash       string
	Channel             *v1alpha1.KafkaTopicChannel
}

// MakeDispatcher generates the dispatcher deployment for the KafKa channel
func MakeDispatcher(args DispatcherArgs) *v1.Deployment {
	dispatcherName = fmt.Sprintf("%s-dispatcher", args.Channel.Name)
	replicas := args.Replicas

	return &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployments",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dispatcherName,
			Namespace: args.DispatcherNamespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Channel),
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: dispatcherLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: dispatcherLabels,
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: args.ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:  DispatcherContainerName,
							Image: args.Image,
							Env:   makeEnv(args),
							Ports: []corev1.ContainerPort{{
								Name:          "metrics",
								ContainerPort: 9090,
							}},
						},
					},
				},
			},
		},
	}
}

func makeEnv(args DispatcherArgs) []corev1.EnvVar {
	env := []corev1.EnvVar{{
		Name:  "KAFKA_BOOTSTRAP_SERVERS",
		Value: strings.Join(args.Channel.Spec.BootstrapServers, ","),
	}, {
		Name:  "KAFKA_TOPIC",
		Value: args.Channel.Spec.Topic,
	}, {
		Name:  "CHANNEL_NAME",
		Value: args.Channel.Name,
	}, {
		Name:  "KAFKA_NET_SASL_ENABLE",
		Value: strconv.FormatBool(args.Channel.Spec.Net.SASL.Enable),
	}, {
		Name:  "KAFKA_NET_TLS_ENABLE",
		Value: strconv.FormatBool(args.Channel.Spec.Net.TLS.Enable),
	}, {
		Name:  system.NamespaceEnvKey,
		Value: args.Channel.Namespace,
	}, {
		Name:  "METRICS_DOMAIN",
		Value: "knative.dev/eventing",
	}, {
		Name:  "CONFIG_LOGGING_NAME",
		Value: "config-logging",
	}, {
		Name:  "CONFIG_LEADERELECTION_NAME",
		Value: "config-leader-election",
	}}

	env = appendEnvFromSecretKeyRef(env, "KAFKA_NET_SASL_USER", args.Channel.Spec.Net.SASL.User.SecretKeyRef)
	env = appendEnvFromSecretKeyRef(env, "KAFKA_NET_SASL_PASSWORD", args.Channel.Spec.Net.SASL.Password.SecretKeyRef)
	env = appendEnvFromSecretKeyRef(env, "KAFKA_NET_SASL_TYPE", args.Channel.Spec.Net.SASL.Type.SecretKeyRef)
	env = appendEnvFromSecretKeyRef(env, "KAFKA_NET_TLS_CERT", args.Channel.Spec.Net.TLS.Cert.SecretKeyRef)
	env = appendEnvFromSecretKeyRef(env, "KAFKA_NET_TLS_KEY", args.Channel.Spec.Net.TLS.Key.SecretKeyRef)
	env = appendEnvFromSecretKeyRef(env, "KAFKA_NET_TLS_CA_CERT", args.Channel.Spec.Net.TLS.CACert.SecretKeyRef)

	if args.DispatcherScope == "namespace" {
		env = append(env, corev1.EnvVar{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		}, corev1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		}, corev1.EnvVar{
			Name:  "CONTAINER_NAME",
			Value: "dispatcher",
		})
	}

	return env
}

func appendEnvFromSecretKeyRef(env []corev1.EnvVar, key string, ref *corev1.SecretKeySelector) []corev1.EnvVar {
	if ref == nil {
		return env
	}

	env = append(env, corev1.EnvVar{
		Name: key,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: ref,
		},
	})

	return env
}

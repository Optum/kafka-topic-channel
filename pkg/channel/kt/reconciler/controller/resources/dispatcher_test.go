package resources

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bindingsv1beta1 "knative.dev/eventing-kafka/pkg/apis/bindings/v1beta1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/system"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
)

const (
	imageName      = "my-test-image"
	serviceAccount = "kafka-ch-dispatcher"
)

func TestNewNamespaceDispatcher(t *testing.T) {
	chanName := "test-channel"
	os.Setenv(system.NamespaceEnvKey, testNS)
	kc := &v1alpha1.KafkaTopicChannel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1aplha1",
			Kind:       "KafkaTopicChannel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      chanName,
			Namespace: testNS,
		},
		Spec: v1alpha1.KafkaTopicChannelSpec{
			Topic: "topic1",
			KafkaAuthSpec: bindingsv1beta1.KafkaAuthSpec{
				BootstrapServers: []string{"server1,server2"},
				Net: bindingsv1beta1.KafkaNetSpec{
					SASL: bindingsv1beta1.KafkaSASLSpec{
						Enable: true,
						User: bindingsv1beta1.SecretValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "the-user-secret",
								},
								Key: "user",
							},
						},
						Password: bindingsv1beta1.SecretValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "the-password-secret",
								},
								Key: "password",
							},
						},
						Type: bindingsv1beta1.SecretValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "the-saslType-secret",
								},
								Key: "saslType",
							},
						},
					},
					TLS: bindingsv1beta1.KafkaTLSSpec{
						Enable: true,
						Cert: bindingsv1beta1.SecretValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "the-cert-secret",
								},
								Key: "tls.crt",
							},
						},
						Key: bindingsv1beta1.SecretValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "the-key-secret",
								},
								Key: "tls.key",
							},
						},
						CACert: bindingsv1beta1.SecretValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "the-ca-cert-secret",
								},
								Key: "tls.crt",
							},
						},
					},
				},
			},
		},
	}

	args := DispatcherArgs{
		DispatcherScope:     "namespace",
		DispatcherNamespace: testNS,
		Image:               imageName,
		Replicas:            1,
		ServiceAccount:      serviceAccount,
		Channel:             kc,
	}

	replicas := int32(1)
	want := &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployments",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNS,
			Name:      fmt.Sprintf("%s-dispatcher", args.Channel.Name),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(kc),
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
					ServiceAccountName: serviceAccount,
					Containers: []corev1.Container{
						{
							Name:  "dispatcher",
							Image: imageName,
							Env: []corev1.EnvVar{{
								Name:  "KAFKA_BOOTSTRAP_SERVERS",
								Value: "server1,server2",
							}, {
								Name:  "KAFKA_TOPIC",
								Value: "topic1",
							}, {
								Name:  "CHANNEL_NAME",
								Value: chanName,
							}, {
								Name:  "KAFKA_NET_SASL_ENABLE",
								Value: "true",
							}, {
								Name:  "KAFKA_NET_TLS_ENABLE",
								Value: "true",
							}, {
								Name:  system.NamespaceEnvKey,
								Value: testNS,
							}, {
								Name:  "METRICS_DOMAIN",
								Value: "knative.dev/eventing",
							}, {
								Name:  "CONFIG_LOGGING_NAME",
								Value: "config-logging",
							}, {
								Name:  "CONFIG_LEADERELECTION_NAME",
								Value: "config-leader-election",
							}, {
								Name: "KAFKA_NET_SASL_USER",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "the-user-secret",
										},
										Key: "user",
									},
								},
							},
								{
									Name: "KAFKA_NET_SASL_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-password-secret",
											},
											Key: "password",
										},
									},
								},
								{
									Name: "KAFKA_NET_SASL_TYPE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-saslType-secret",
											},
											Key: "saslType",
										},
									},
								},
								{
									Name: "KAFKA_NET_TLS_CERT",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-cert-secret",
											},
											Key: "tls.crt",
										},
									},
								},
								{
									Name: "KAFKA_NET_TLS_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-key-secret",
											},
											Key: "tls.key",
										},
									},
								},
								{
									Name: "KAFKA_NET_TLS_CA_CERT",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-ca-cert-secret",
											},
											Key: "tls.crt",
										},
									},
								}, {
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								}, {
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								}, {
									Name:  "CONTAINER_NAME",
									Value: "dispatcher",
								}},
							Ports: []corev1.ContainerPort{{
								Name:          "metrics",
								ContainerPort: 9090,
							}},
						}},
				},
			},
		},
	}

	got := MakeDispatcher(args)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected condition (-want, +got) = %v", diff)
	}
}

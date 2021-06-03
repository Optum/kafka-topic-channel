package resources

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
)

const (
	kcName             = "my-test-kc"
	testNS             = "my-test-ns"
	testDispatcherNS   = "dispatcher-namespace"
	testDispatcherName = "dispatcher-name"
	testConfigMapHash  = "deadbeef"
)

func TestMakeChannelServiceAddress(t *testing.T) {
	if want, got := "my-test-kc-kn-channel", MakeChannelServiceName(kcName); want != got {
		t.Errorf("Want: %q got %q", want, got)
	}
}

func TestMakeService(t *testing.T) {
	imc := &v1alpha1.KafkaTopicChannel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1aplha1",
			Kind:       "KafkaTopicChannel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kcName,
			Namespace: testNS,
		},
	}
	want := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kn-channel", kcName),
			Namespace: testNS,
			Labels: map[string]string{
				MessagingRoleLabel: MessagingRole,
			},
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(imc),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     portName,
					Protocol: corev1.ProtocolTCP,
					Port:     portNumber,
				},
			},
		},
	}

	got, err := MakeK8sService(imc)
	if err != nil {
		t.Fatalf("Failed to create new service: %s", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected condition (-want, +got) = %v", diff)
	}
}

func TestMakeServiceWithExternal(t *testing.T) {
	imc := &v1alpha1.KafkaTopicChannel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kcName,
			Namespace: testNS,
		},
	}
	want := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kn-channel", kcName),
			Namespace: testNS,
			Labels: map[string]string{
				MessagingRoleLabel: MessagingRole,
			},
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(imc),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: "dispatcher-name.dispatcher-namespace.svc.cluster.local",
		},
	}

	got, err := MakeK8sService(imc, ExternalService(testDispatcherNS, testDispatcherName))
	if err != nil {
		t.Fatalf("Failed to create new service: %s", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected condition (-want, +got) = %v", diff)
	}
}

func TestMakeServiceWithFailingOption(t *testing.T) {
	imc := &v1alpha1.KafkaTopicChannel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kcName,
			Namespace: testNS,
		},
	}
	_, err := MakeK8sService(imc, func(svc *corev1.Service) error { return errors.New("test-induced failure") })
	if err == nil {
		t.Fatalf("Expcted error from new service but got none")
	}
}

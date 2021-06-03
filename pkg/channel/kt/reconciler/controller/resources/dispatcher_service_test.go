package resources

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
)

func TestNewDispatcherService(t *testing.T) {
	kc := &v1alpha1.KafkaTopicChannel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1aplha1",
			Kind:       "KafkaTopicChannel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-channel",
			Namespace: testNS,
		},
	}
	want := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-dispatcher", kc.Name),
			Namespace: testNS,
			Labels:    dispatcherLabels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(kc),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: dispatcherLabels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http-dispatcher",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: 8080},
				},
			},
		},
	}

	got := MakeDispatcherService(testNS, kc)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected condition (-want, +got) = %v", diff)
	}
}

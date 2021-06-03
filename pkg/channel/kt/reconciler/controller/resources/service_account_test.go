package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewServiceAccount(t *testing.T) {
	want := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNS,
			Name:      serviceAccount,
		},
	}

	got := MakeServiceAccount(testNS, serviceAccount)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected condition (-want, +got) = %v", diff)
	}
}

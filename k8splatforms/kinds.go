package k8splatforms

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KindProcessor interface {
	Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error)
	IsActive(obj client.Object) bool
	VirtualPods(obj client.Object) []VirtualPod
}

type VirtualPod struct {
	metav1.ObjectMeta
	Spec    corev1.PodSpec
	SubName string
}

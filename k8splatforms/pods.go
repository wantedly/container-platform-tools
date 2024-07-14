package k8splatforms

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodProcessor struct{}

var _ KindProcessor = PodProcessor{}

// Retrieve implements KindProcessor.
func (p PodProcessor) Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error) {
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}
	objs := make([]client.Object, len(pods.Items))
	for i := range pods.Items {
		objs[i] = &pods.Items[i]
	}
	return objs, nil
}

// IsActive implements KindProcessor.
func (p PodProcessor) IsActive(obj client.Object) bool {
	if pod, ok := obj.(*corev1.Pod); ok {
		return !podFinishedOrDeleted(pod)
	}
	return false
}

func podFinishedOrDeleted(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.ObjectMeta.DeletionTimestamp != nil
}

// VirtualPods implements KindProcessor.
func (p PodProcessor) VirtualPods(obj client.Object) []corev1.PodTemplateSpec {
	if pod, ok := obj.(*corev1.Pod); ok {
		return []corev1.PodTemplateSpec{
			{
				ObjectMeta: pod.ObjectMeta,
				Spec:       pod.Spec,
			},
		}
	}
	return nil
}

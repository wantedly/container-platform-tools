package k8splatforms

import (
	"context"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DaemonSetProcessor struct{}

var _ KindProcessor = DaemonSetProcessor{}

// Retrieve implements KindProcessor.
func (p DaemonSetProcessor) Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error) {
	daemonSets, err := clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list daemon sets")
	}
	objs := make([]client.Object, len(daemonSets.Items))
	for i := range daemonSets.Items {
		objs[i] = &daemonSets.Items[i]
	}
	return objs, nil
}

// IsRelevant implements KindProcessor.
func (p DaemonSetProcessor) IsRelevant(obj client.Object) bool {
	if _, ok := obj.(*appsv1.DaemonSet); ok {
		return true
	}
	return false
}

// VirtualPods implements KindProcessor.
func (p DaemonSetProcessor) VirtualPods(obj client.Object) []corev1.PodTemplateSpec {
	if daemonSet, ok := obj.(*appsv1.DaemonSet); ok {
		return []corev1.PodTemplateSpec{
			daemonSet.Spec.Template,
		}
	}
	return nil
}

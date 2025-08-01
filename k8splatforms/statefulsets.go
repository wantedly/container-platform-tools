package k8splatforms

import (
	"context"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSetProcessor struct{}

var _ KindProcessor = StatefulSetProcessor{}

// Retrieve implements KindProcessor.
func (p StatefulSetProcessor) Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error) {
	statefulSets, err := clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list stateful sets")
	}
	objs := make([]client.Object, len(statefulSets.Items))
	for i := range statefulSets.Items {
		objs[i] = &statefulSets.Items[i]
	}
	return objs, nil
}

// IsActive implements KindProcessor.
func (p StatefulSetProcessor) IsActive(obj client.Object) bool {
	if _, ok := obj.(*appsv1.StatefulSet); ok {
		return true
	}
	return false
}

// VirtualPods implements KindProcessor.
func (p StatefulSetProcessor) VirtualPods(obj client.Object) []VirtualPod {
	if statefulSet, ok := obj.(*appsv1.StatefulSet); ok {
		return []VirtualPod{
			{
				ObjectMeta: statefulSet.Spec.Template.ObjectMeta,
				Spec:       statefulSet.Spec.Template.Spec,
			},
		}
	}
	return nil
}

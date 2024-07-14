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

type ReplicaSetProcessor struct{}

var _ KindProcessor = ReplicaSetProcessor{}

// Retrieve implements KindProcessor.
func (p ReplicaSetProcessor) Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error) {
	replicaSets, err := clientset.AppsV1().ReplicaSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list replica sets")
	}
	objs := make([]client.Object, len(replicaSets.Items))
	for i := range replicaSets.Items {
		objs[i] = &replicaSets.Items[i]
	}
	return objs, nil
}

// IsActive implements KindProcessor.
func (p ReplicaSetProcessor) IsActive(obj client.Object) bool {
	if rs, ok := obj.(*appsv1.ReplicaSet); ok {
		return replicaSetReplicas(rs) > 0
	}
	return false
}

func replicaSetReplicas(rs *appsv1.ReplicaSet) int32 {
	if rs.Spec.Replicas == nil {
		return 1
	}
	return *rs.Spec.Replicas
}

// VirtualPods implements KindProcessor.
func (p ReplicaSetProcessor) VirtualPods(obj client.Object) []VirtualPod {
	if rs, ok := obj.(*appsv1.ReplicaSet); ok {
		return []VirtualPod{
			{
				ObjectMeta: rs.Spec.Template.ObjectMeta,
				Spec:       rs.Spec.Template.Spec,
			},
		}
	}
	return nil
}

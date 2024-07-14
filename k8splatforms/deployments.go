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

type DeploymentProcessor struct{}

var _ KindProcessor = DeploymentProcessor{}

// Retrieve implements KindProcessor.
func (p DeploymentProcessor) Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error) {
	deployments, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list deployments")
	}
	objs := make([]client.Object, len(deployments.Items))
	for i := range deployments.Items {
		objs[i] = &deployments.Items[i]
	}
	return objs, nil
}

// IsActive implements KindProcessor.
func (p DeploymentProcessor) IsActive(obj client.Object) bool {
	if _, ok := obj.(*appsv1.Deployment); ok {
		return true
	}
	return false
}

// VirtualPods implements KindProcessor.
func (p DeploymentProcessor) VirtualPods(obj client.Object) []corev1.PodTemplateSpec {
	if deployment, ok := obj.(*appsv1.Deployment); ok {
		return []corev1.PodTemplateSpec{
			deployment.Spec.Template,
		}
	}
	return nil
}

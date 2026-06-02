package k8splatforms

import (
	"context"

	workflowv1alpha1 "github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1"
	versioned "github.com/argoproj/argo-workflows/v4/pkg/client/clientset/versioned"
	argoscheme "github.com/argoproj/argo-workflows/v4/pkg/client/clientset/versioned/scheme"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronWorkflowProcessor struct{}

var _ KindProcessor = CronWorkflowProcessor{}

// Retrieve implements KindProcessor.
func (c CronWorkflowProcessor) Retrieve(ctx context.Context, config *rest.Config, _clientset *kubernetes.Clientset) ([]client.Object, error) {
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create argo clientset")
	}
	cronWorkflows, err := clientset.ArgoprojV1alpha1().CronWorkflows("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list cron workflows")
	}
	objs := make([]client.Object, len(cronWorkflows.Items))
	for i := range cronWorkflows.Items {
		objs[i] = &cronWorkflows.Items[i]
	}
	return objs, nil
}

// Scheme implements KindProcessor.
func (c CronWorkflowProcessor) Scheme() *runtime.Scheme {
	return argoscheme.Scheme
}

// IsActive implements KindProcessor.
func (c CronWorkflowProcessor) IsActive(obj client.Object) bool {
	if _, ok := obj.(*workflowv1alpha1.CronWorkflow); ok {
		return true
	}
	return false
}

// VirtualPods implements KindProcessor.
func (c CronWorkflowProcessor) VirtualPods(obj client.Object) []VirtualPod {
	if cronWorkflow, ok := obj.(*workflowv1alpha1.CronWorkflow); ok {
		return collectWorkflowPods(cronWorkflow.Spec.WorkflowSpec)
	}
	return nil
}

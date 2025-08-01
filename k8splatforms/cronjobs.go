package k8splatforms

import (
	"context"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronJobProcessor struct{}

var _ KindProcessor = CronJobProcessor{}

// Retrieve implements KindProcessor.
func (c CronJobProcessor) Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error) {
	cronJobs, err := clientset.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list cron jobs")
	}
	objs := make([]client.Object, len(cronJobs.Items))
	for i := range cronJobs.Items {
		objs[i] = &cronJobs.Items[i]
	}
	return objs, nil
}

// IsActive implements KindProcessor.
func (c CronJobProcessor) IsActive(obj client.Object) bool {
	if _, ok := obj.(*batchv1.CronJob); ok {
		return true
	}
	return false
}

// VirtualPods implements KindProcessor.
func (c CronJobProcessor) VirtualPods(obj client.Object) []VirtualPod {
	if cronJob, ok := obj.(*batchv1.CronJob); ok {
		return []VirtualPod{
			{
				ObjectMeta: cronJob.Spec.JobTemplate.Spec.Template.ObjectMeta,
				Spec:       cronJob.Spec.JobTemplate.Spec.Template.Spec,
			},
		}
	}
	return nil
}

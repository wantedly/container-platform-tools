package k8splatforms

import (
	"context"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobProcessor struct{}

var _ KindProcessor = JobProcessor{}

// Retrieve implements KindProcessor.
func (p JobProcessor) Retrieve(ctx context.Context, config *rest.Config, clientset *kubernetes.Clientset) ([]client.Object, error) {
	jobs, err := clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list jobs")
	}
	objs := make([]client.Object, len(jobs.Items))
	for i := range jobs.Items {
		objs[i] = &jobs.Items[i]
	}
	return objs, nil
}

// IsActive implements KindProcessor.
func (p JobProcessor) IsActive(obj client.Object) bool {
	if job, ok := obj.(*batchv1.Job); ok {
		return !hasJobFinished(job)
	}
	return false
}

func hasJobFinished(job *batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if condition.Status == corev1.ConditionTrue && (condition.Type == batchv1.JobComplete || condition.Type == batchv1.JobFailed) {
			return true
		}
	}
	return false
}

// VirtualPods implements KindProcessor.
func (p JobProcessor) VirtualPods(obj client.Object) []VirtualPod {
	if job, ok := obj.(*batchv1.Job); ok {
		return []VirtualPod{
			{
				ObjectMeta: job.Spec.Template.ObjectMeta,
				Spec:       job.Spec.Template.Spec,
			},
		}
	}
	return nil
}

package k8splatforms

import (
	"context"

	workflowv1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	versioned "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkflowProcessor struct{}

var _ KindProcessor = WorkflowProcessor{}

// Retrieve implements KindProcessor.
func (w WorkflowProcessor) Retrieve(ctx context.Context, config *rest.Config, _clientset *kubernetes.Clientset) ([]client.Object, error) {
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create argo clientset")
	}
	workflows, err := clientset.ArgoprojV1alpha1().Workflows("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list cron workflows")
	}
	objs := make([]client.Object, len(workflows.Items))
	for i := range workflows.Items {
		objs[i] = &workflows.Items[i]
	}
	return objs, nil
}

// IsActive implements KindProcessor.
func (w WorkflowProcessor) IsActive(obj client.Object) bool {
	if _, ok := obj.(*workflowv1alpha1.Workflow); ok {
		return true
	}
	return false
}

// VirtualPods implements KindProcessor.
func (w WorkflowProcessor) VirtualPods(obj client.Object) []corev1.PodTemplateSpec {
	if workflow, ok := obj.(*workflowv1alpha1.Workflow); ok {
		return collectWorkflowPods(workflow.Spec)
	}
	return nil
}

func collectWorkflowPods(spec workflowv1alpha1.WorkflowSpec) []corev1.PodTemplateSpec {
	var pods []corev1.PodTemplateSpec
	for _, template := range spec.Templates {
		collectWorkflowPodsImpl(spec, template, &pods)
	}
	return pods
}
func collectWorkflowPodsImpl(spec workflowv1alpha1.WorkflowSpec, template workflowv1alpha1.Template, pods *[]corev1.PodTemplateSpec) {
	// https://github.com/argoproj/argo-workflows/blob/v3.5.8/workflow/controller/workflowpod.go#L78

	if template.Container != nil || template.Script != nil {
		var containers []corev1.Container
		if template.Container != nil {
			containers = []corev1.Container{*template.Container}
		} else if template.ContainerSet != nil {
			for _, container := range template.ContainerSet.Containers {
				containers = append(containers, container.Container)
			}
		} else if template.Script != nil {
			containers = []corev1.Container{template.Script.Container}
		}
		for _, sidecar := range template.Sidecars {
			containers = append(containers, sidecar.Container)
		}

		var initContainers []corev1.Container
		for _, initContainer := range template.InitContainers {
			initContainers = append(initContainers, initContainer.Container)
		}

		var nodeSelector map[string]string
		if len(template.NodeSelector) > 0 {
			nodeSelector = template.NodeSelector
		} else if len(spec.NodeSelector) > 0 {
			nodeSelector = spec.NodeSelector
		}

		var affinity *corev1.Affinity
		if template.Affinity != nil {
			affinity = template.Affinity
		} else if spec.Affinity != nil {
			affinity = spec.Affinity
		}

		var tolerations []corev1.Toleration
		if len(template.Tolerations) > 0 {
			tolerations = template.Tolerations
		} else if len(spec.Tolerations) > 0 {
			tolerations = spec.Tolerations
		}

		*pods = append(*pods, corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				InitContainers: initContainers,
				Containers:     containers,
				NodeSelector:   nodeSelector,
				Affinity:       affinity,
				Tolerations:    tolerations,
			},
		})
	} else if template.Steps != nil {
		for _, parallelStep := range template.Steps {
			for _, step := range parallelStep.Steps {
				if step.Inline != nil {
					collectWorkflowPodsImpl(spec, *step.Inline, pods)
				}
			}
		}
	} else if template.DAG != nil {
		for _, task := range template.DAG.Tasks {
			if task.Inline != nil {
				collectWorkflowPodsImpl(spec, *task.Inline, pods)
			}
		}
	}
}

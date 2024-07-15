package k8splatforms

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	corev1 "k8s.io/api/core/v1"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Row struct {
	Namespace            string
	APIVersion           string
	Kind                 string
	Name                 string
	SubName              string
	ScheduledPlatform    *dockerplatforms.DockerPlatform
	DeclaredPlatforms    dockerplatforms.DockerPlatformList
	ImagePlatforms       dockerplatforms.DockerPlatformList
	ImagePlatformDetails map[string]dockerplatforms.DockerPlatformList
	HasViolation         bool
	CPUUsage             float64
	MemoryUsage          float64
	Error                string
}

func EvaluateObjects(
	ctx context.Context,
	objs []client.Object,
	nodes []corev1.Node,
	metricses []metricsv1beta1.PodMetrics,
	after time.Time,
	nodePlatforms dockerplatforms.DockerPlatformList,
	platformInspector dockerplatforms.PlatformInspector,
	processors []KindProcessor,
) ([]Row, errorutil.Aggregate) {
	nodePlatforms = nodePlatforms.Variantless()
	nodesByName := make(map[string]*corev1.Node)
	for _, node := range nodes {
		nodesByName[node.Name] = &node
	}
	metricsesByName := make(map[string]*metricsv1beta1.PodMetrics)
	for _, metrics := range metricses {
		metricsesByName[metrics.Namespace+"/"+metrics.Name] = &metrics
	}

	var rows []Row
	var errs []error
	for _, obj := range objs {
		var virtualPods []VirtualPod
		for _, processor := range processors {
			if obj.GetCreationTimestamp().After(after) || processor.IsActive(obj) {
				virtualPods = append(virtualPods, processor.VirtualPods(obj)...)
				break
			}
		}
		for _, virtualPod := range virtualPods {
			row, err := evaluateVirtualPod(ctx, obj, nodesByName, metricsesByName, nodePlatforms, platformInspector, virtualPod)
			if err != nil {
				for _, err := range err.Errors() {
					errs = append(errs, errors.Wrap(err, "evaluating pod platforms"))
				}
			}
			rows = append(rows, row)
		}
	}
	if len(errs) > 0 {
		return rows, errorutil.NewAggregate(errs)
	}
	return rows, nil
}

func evaluateVirtualPod(
	ctx context.Context,
	obj client.Object,
	nodesByName map[string]*corev1.Node,
	metricsesByName map[string]*metricsv1beta1.PodMetrics,
	nodePlatforms dockerplatforms.DockerPlatformList,
	platformInspector dockerplatforms.PlatformInspector,
	virtualPod VirtualPod,
) (Row, errorutil.Aggregate) {
	var scheduledPlatform *dockerplatforms.DockerPlatform
	var cpuUsage float64
	var memoryUsage float64
	if pod, ok := obj.(*corev1.Pod); ok {
		if node, ok := nodesByName[pod.Spec.NodeName]; ok {
			scheduledPlatform = &dockerplatforms.DockerPlatform{
				OS:           node.ObjectMeta.Labels["kubernetes.io/os"],
				Architecture: node.ObjectMeta.Labels["kubernetes.io/arch"],
			}
		}
		if metrics, ok := metricsesByName[pod.Namespace+"/"+pod.Name]; ok {
			for _, container := range metrics.Containers {
				if cpuQuant, ok := container.Usage[corev1.ResourceCPU]; ok {
					cpuUsage += cpuQuant.AsApproximateFloat64()
				}
				if memoryQuant, ok := container.Usage[corev1.ResourceMemory]; ok {
					memoryUsage += memoryQuant.AsApproximateFloat64()
				}
			}
		}
	}
	declaredPlatforms := PodPlatforms(&corev1.Pod{
		ObjectMeta: virtualPod.ObjectMeta,
		Spec:       virtualPod.Spec,
	}, nodePlatforms)
	imagePlatformDetails := make(map[string]dockerplatforms.DockerPlatformList)
	var imagePlatforms dockerplatforms.DockerPlatformList
	found := false
	var errs []error
	for _, container := range virtualPod.Spec.Containers {
		platforms, err := platformInspector.GetPlatforms(ctx, container.Image)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "inspecting image platforms"))
			continue
		}
		platforms2 := dockerplatforms.DockerPlatformList(platforms).Variantless()
		imagePlatformDetails[container.Name] = platforms2
		if found {
			imagePlatforms = imagePlatforms.Intersection(platforms2)
		} else {
			imagePlatforms = platforms2
			found = true
		}
	}
	if !found {
		imagePlatforms = nodePlatforms
	}
	row := Row{
		Namespace:            obj.GetNamespace(),
		APIVersion:           obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:                 obj.GetObjectKind().GroupVersionKind().Kind,
		Name:                 obj.GetName(),
		SubName:              virtualPod.SubName,
		ScheduledPlatform:    scheduledPlatform,
		DeclaredPlatforms:    declaredPlatforms,
		ImagePlatforms:       imagePlatforms,
		ImagePlatformDetails: imagePlatformDetails,
		HasViolation:         !imagePlatforms.Includes(declaredPlatforms),
		CPUUsage:             cpuUsage,
		MemoryUsage:          memoryUsage,
	}
	if len(errs) > 0 {
		row.Error = errs[0].Error()
		return row, errorutil.NewAggregate(errs)
	}
	return row, nil
}

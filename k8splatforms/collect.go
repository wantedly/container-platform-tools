package k8splatforms

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Collector struct {
	RESTConfig        *rest.Config
	After             time.Time
	NodePlatforms     dockerplatforms.DockerPlatformList
	PlatformInspector dockerplatforms.PlatformInspector
	Processors        []KindProcessor
}

func (c Collector) Collect(
	ctx context.Context,
) ([]Row, error) {
	clientset, err := kubernetes.NewForConfig(c.RESTConfig)
	if err != nil {
		return nil, errors.Wrap(err, "creating clientset")
	}

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list nodes")
	}

	mClientset, err := versioned.NewForConfig(c.RESTConfig)
	if err != nil {
		return nil, errors.Wrap(err, "creating clientset for metrics")
	}
	metricses, err := mClientset.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pod metricses")
	}

	var objs []client.Object
	for _, processor := range c.Processors {
		processorObjs, err := processor.Retrieve(ctx, c.RESTConfig, clientset)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve objects")
		}
		objs = append(objs, processorObjs...)
	}
	objs = SortObjects(objs)

	return EvaluateObjects(
		ctx,
		objs,
		nodes.Items,
		metricses.Items,
		c.After,
		c.NodePlatforms,
		c.PlatformInspector,
		c.Processors,
	)
}

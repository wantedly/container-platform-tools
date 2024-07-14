package k8splatforms

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Collect(
	ctx context.Context,
	config *rest.Config,
	clientset *kubernetes.Clientset,
	processors []KindProcessor,
) ([]client.Object, error) {
	// nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	// if err != nil {
	// 	return errors.Wrap(err, "failed to list cron jobs")
	// }

	var objs []client.Object
	for _, processor := range processors {
		processorObjs, err := processor.Retrieve(ctx, config, clientset)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve objects")
		}
		objs = append(objs, processorObjs...)
	}

	// filteredObjs := make([]client.Object, 0, len(objs))
	// for _, obj := range objs {
	// 	relevant := false
	// 	if obj.GetCreationTimestamp().After(after) {
	// 		relevant = true
	// 	}
	// 	for _, processor := range processors {
	// 		if processor.IsRelevant(obj) {
	// 			relevant = true
	// 			break
	// 		}
	// 	}
	// 	if relevant {
	// 		filteredObjs = append(filteredObjs, obj)
	// 	}
	// }

	return SortObjects(objs), nil
}

package k8splatforms

import (
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SortObjects(
	objs []client.Object,
) []client.Object {
	// Make it deterministic in the presence of cycles (which should not happen) or duplicates
	slices.SortFunc(objs, objectCmp)
	forest := makeForest(objs)
	return flattenForest(forest)
}

func objectsByKey(objs []client.Object) map[objectKey]client.Object {
	objByKey := make(map[objectKey]client.Object)
	for _, obj := range objs {
		objByKey[getObjectKey(obj)] = obj
	}
	return objByKey
}

type objectForest struct {
	Roots    []client.Object
	Children map[client.Object][]client.Object
}

func makeForest(objs []client.Object) objectForest {
	byKey := objectsByKey(objs)
	forest := objectForest{
		Roots:    make([]client.Object, 0, len(objs)),
		Children: make(map[client.Object][]client.Object),
	}
	visited := make(map[client.Object]bool)
	for _, obj := range objs {
		makeForestRecursion(obj, byKey, &forest, visited)
	}
	return forest
}

func makeForestRecursion(obj client.Object, byKey map[objectKey]client.Object, forest *objectForest, visited map[client.Object]bool) bool {
	if visPost, ok := visited[obj]; ok {
		return visPost
	}
	visited[obj] = false
	for _, ownerRef := range obj.GetOwnerReferences() {
		ownerObj, ok := byKey[getOwnerRefKey(obj.GetNamespace(), ownerRef)]
		if !ok {
			continue
		}
		if makeForestRecursion(ownerObj, byKey, forest, visited) {
			forest.Children[ownerObj] = append(forest.Children[ownerObj], obj)
			visited[obj] = true
			return true
		}
	}
	forest.Roots = append(forest.Roots, obj)
	visited[obj] = true
	return true
}

func flattenForest(forest objectForest) []client.Object {
	var objs []client.Object
	slices.SortFunc(forest.Roots, objectCmp)
	for _, root := range forest.Roots {
		objs = append(objs, root)
		flattenForestRecursion(root, forest, &objs)
	}
	return objs
}

func flattenForestRecursion(obj client.Object, forest objectForest, objs *[]client.Object) {
	children := forest.Children[obj]
	slices.SortFunc(children, objectCmp)
	for _, child := range children {
		*objs = append(*objs, child)
		flattenForestRecursion(child, forest, objs)
	}
}

func objectCmp(a, b client.Object) int {
	nsA := a.GetNamespace()
	nsB := b.GetNamespace()
	if nsA < nsB {
		return -1
	} else if nsA > nsB {
		return 1
	}
	gvkA := a.GetObjectKind().GroupVersionKind()
	gvkB := b.GetObjectKind().GroupVersionKind()
	gvA := gvkA.GroupVersion()
	gvB := gvkB.GroupVersion()
	if gvA.String() < gvB.String() {
		return -1
	} else if gvA.String() > gvB.String() {
		return 1
	}
	if gvkA.Kind < gvkB.Kind {
		return -1
	} else if gvkA.Kind > gvkB.Kind {
		return 1
	}
	if a.GetName() < b.GetName() {
		return -1
	} else if a.GetName() > b.GetName() {
		return 1
	}
	return 0
}

type objectKey struct {
	Namespace  string
	APIVersion string
	Kind       string
	Name       string
}

func getObjectKey(obj client.Object) objectKey {
	return objectKey{
		Namespace:  obj.GetNamespace(),
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		Name:       obj.GetName(),
	}
}

func getOwnerRefKey(ns string, ownerRef metav1.OwnerReference) objectKey {
	return objectKey{
		Namespace:  ns,
		APIVersion: ownerRef.APIVersion,
		Kind:       ownerRef.Kind,
		Name:       ownerRef.Name,
	}
}

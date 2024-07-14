package k8splatforms

import (
	"slices"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SortObjects(
	objs []client.Object,
) []client.Object {
	graph := buildGraph(objs)
	graph.sortObjects()
	return graph.objects()
}

type objectGraph struct {
	ByName     map[string]client.Object
	Names      []string
	OwningRefs map[string][]string
	OwnerRefs  map[string][]string
}

func buildGraph(objs []client.Object) objectGraph {
	objByName := make(map[string]client.Object)
	objNames := make([]string, 0, len(objs))
	owningRefs := make(map[string][]string)
	ownerRefs := make(map[string][]string)
	for _, obj := range objs {
		objKey := objectKey(obj)
		objByName[objKey] = obj
		objNames = append(objNames, objKey)
	}
	for _, obj := range objs {
		objKey := objectKey(obj)
		for _, ownerRef := range obj.GetOwnerReferences() {
			ownerObj := unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": ownerRef.APIVersion,
					"kind":       ownerRef.Kind,
					"metadata": map[string]interface{}{
						"name":      ownerRef.Name,
						"namespace": obj.GetNamespace(),
					},
				},
			}
			ownerKey := objectKey(&ownerObj)
			if _, ok := objByName[ownerKey]; ok {
				ownerRefs[objKey] = append(ownerRefs[objKey], ownerKey)
				owningRefs[ownerKey] = append(owningRefs[ownerKey], objKey)
			}
		}
	}
	return objectGraph{
		ByName:     objByName,
		Names:      objNames,
		OwningRefs: owningRefs,
		OwnerRefs:  ownerRefs,
	}
}

func (g *objectGraph) sortObjects() {
	connectedComponents := g.connectedComponents()
	var newObjNames []string
	for _, cc := range connectedComponents {
		slices.SortFunc(cc, func(a, b string) int {
			return objectCmp(g.ByName[a], g.ByName[b])
		})
		cc = g.topologicalSort(cc)
		newObjNames = append(newObjNames, cc...)
	}
	g.Names = newObjNames
}

func (g *objectGraph) connectedComponents() [][]string {
	var connectedComponents [][]string
	inCC := make(map[string]struct{})
	for _, objName := range g.Names {
		if _, ok := inCC[objName]; ok {
			continue
		}
		var cc []string
		g.findCC(objName, &cc, inCC)
		connectedComponents = append(connectedComponents, cc)
	}
	return connectedComponents
}
func (g *objectGraph) findCC(objName string, cc *[]string, inCC map[string]struct{}) {
	inCC[objName] = struct{}{}
	*cc = append(*cc, objName)
	for _, ownerRef := range g.OwningRefs[objName] {
		if _, ok := inCC[ownerRef]; ok {
			continue
		}
		g.findCC(ownerRef, cc, inCC)
	}
	for _, ownerRef := range g.OwnerRefs[objName] {
		if _, ok := inCC[ownerRef]; ok {
			continue
		}
		g.findCC(ownerRef, cc, inCC)
	}
}

func (g *objectGraph) topologicalSort(objNames []string) []string {
	newObjNames := make([]string, 0, len(objNames))
	visited := make(map[string]struct{})
	for _, objName := range objNames {
		if _, ok := visited[objName]; ok {
			continue
		}
		g.topologicalSortVisit(objName, &newObjNames, visited)
	}
	return newObjNames
}
func (g *objectGraph) topologicalSortVisit(objName string, newObjNames *[]string, visited map[string]struct{}) {
	visited[objName] = struct{}{}
	ownerNames := slices.Clone(g.OwnerRefs[objName])
	slices.SortFunc(ownerNames, func(a, b string) int {
		return objectCmp(g.ByName[a], g.ByName[b])
	})
	for _, ownerName := range ownerNames {
		if _, ok := visited[ownerName]; ok {
			continue
		}
		g.topologicalSortVisit(ownerName, newObjNames, visited)
	}
	*newObjNames = append(*newObjNames, objName)
}

func (g *objectGraph) objects() []client.Object {
	objs := make([]client.Object, 0, len(g.Names))
	for _, objName := range g.Names {
		objs = append(objs, g.ByName[objName])
	}
	return objs
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

func objectKey(obj client.Object) string {
	gvk := obj.GetObjectKind().GroupVersionKind()
	return obj.GetNamespace() + ":" + gvk.GroupVersion().String() + "." + gvk.Kind + "/" + obj.GetName()
}

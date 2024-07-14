package k8splatforms_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/k8splatforms"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var podTypeMeta = metav1.TypeMeta{
	APIVersion: "v1",
	Kind:       "Pod",
}
var replicaSetTypeMeta = metav1.TypeMeta{
	APIVersion: "apps/v1",
	Kind:       "ReplicaSet",
}
var deploymentTypeMeta = metav1.TypeMeta{
	APIVersion: "apps/v1",
	Kind:       "Deployment",
}

func TestSortObjects(t *testing.T) {
	testcases := []struct {
		name     string
		objs     []client.Object
		expected []string
	}{
		{
			name:     "no objects",
			objs:     []client.Object{},
			expected: []string{},
		},
		{
			name: "single object",
			objs: []client.Object{
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod1",
					},
				},
			},
			expected: []string{"default:Pod/pod1"},
		},
		{
			name: "deployment-replicaset-pod",
			objs: []client.Object{
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod1",
					},
				},
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod2",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "ReplicaSet",
								Name:       "rs3",
							},
						},
					},
				},
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod3",
					},
				},
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod4",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "ReplicaSet",
								Name:       "rs1",
							},
						},
					},
				},
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod5",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "ReplicaSet",
								Name:       "rs3",
							},
						},
					},
				},
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod6",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "ReplicaSet",
								Name:       "rs2",
							},
						},
					},
				},
				&corev1.Pod{
					TypeMeta: podTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "pod7",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "ReplicaSet",
								Name:       "rs1",
							},
						},
					},
				},
				&appsv1.ReplicaSet{
					TypeMeta: replicaSetTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "rs1",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "d2",
							},
						},
					},
				},
				&appsv1.ReplicaSet{
					TypeMeta: replicaSetTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "rs2",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "d2",
							},
						},
					},
				},
				&appsv1.ReplicaSet{
					TypeMeta: replicaSetTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "rs3",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "d2",
							},
						},
					},
				},
				&appsv1.Deployment{
					TypeMeta: deploymentTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "d1",
					},
				},
				&appsv1.Deployment{
					TypeMeta: deploymentTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "d2",
					},
				},
				&appsv1.Deployment{
					TypeMeta: deploymentTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "d3",
					},
				},
			},
			expected: []string{
				"default:Deployment/d1",
				"default:Deployment/d2",
				"default:ReplicaSet/rs1",
				"default:Pod/pod4",
				"default:Pod/pod7",
				"default:ReplicaSet/rs2",
				"default:Pod/pod6",
				"default:ReplicaSet/rs3",
				"default:Pod/pod2",
				"default:Pod/pod5",
				"default:Deployment/d3",
				"default:Pod/pod1",
				"default:Pod/pod3",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sortedObjs := k8splatforms.SortObjects(tc.objs)

			sortedObjKeys := make([]string, len(sortedObjs))
			for i, obj := range sortedObjs {
				sortedObjKeys[i] = obj.GetNamespace() + ":" + obj.GetObjectKind().GroupVersionKind().Kind + "/" + obj.GetName()
			}

			if diff := cmp.Diff(tc.expected, sortedObjKeys); diff != "" {
				t.Errorf("unexpected sorted objects (-want +got):\n%s", diff)
			}
		})
	}
}

package k8splatforms_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	"github.com/wantedly/container-platform-tools/k8splatforms"
	corev1 "k8s.io/api/core/v1"
)

func TestPodPlatforms(t *testing.T) {
	testcases := []struct {
		name          string
		pod           *corev1.Pod
		nodePlatforms dockerplatforms.DockerPlatformList
		expected      dockerplatforms.DockerPlatformList
	}{
		{
			name:          "empty",
			pod:           &corev1.Pod{},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64, linux/arm64, windows/amd64"),
		},
		{
			name: "label selector os",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
				},
			},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64, linux/arm64"),
		},
		{
			name: "label selector arch",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"kubernetes.io/arch": "amd64",
					},
				},
			},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64, windows/amd64"),
		},
		{
			name: "label selector os/arch",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"kubernetes.io/os":   "linux",
						"kubernetes.io/arch": "amd64",
					},
				},
			},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64"),
		},
		{
			name: "node affinity os",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/os",
												Operator: "In",
												Values:   []string{"linux"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64, linux/arm64"),
		},
		{
			name: "node affinity arch",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/arch",
												Operator: "In",
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64, windows/amd64"),
		},
		{
			name: "node affinity os/arch",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/os",
												Operator: "In",
												Values:   []string{"linux"},
											},
											{
												Key:      "kubernetes.io/arch",
												Operator: "In",
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64"),
		},
		{
			name: "mixed label selector and node affinity",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/arch",
												Operator: "In",
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			nodePlatforms: pl(t, "linux/amd64, linux/arm64, windows/amd64"),
			expected:      pl(t, "linux/amd64"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := k8splatforms.PodPlatforms(tc.pod, tc.nodePlatforms)
			if diff := cmp.Diff(tc.expected.String(), actual.String()); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func pl(t *testing.T, s string) dockerplatforms.DockerPlatformList {
	t.Helper()

	list, err := dockerplatforms.ParseDockerPlatformList(s)
	if err != nil {
		t.Fatal(err)
	}
	return list
}

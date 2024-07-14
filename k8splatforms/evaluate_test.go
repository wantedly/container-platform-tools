package k8splatforms_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	dockerplatformstesting "github.com/wantedly/container-platform-tools/dockerplatforms/testing"
	"github.com/wantedly/container-platform-tools/k8splatforms"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var time1 = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

func TestEvaluateObjects(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	nodePlatforms := dockerplatforms.DockerPlatformList{
		{
			OS:           "linux",
			Architecture: "amd64",
		},
		{
			OS:           "linux",
			Architecture: "arm64",
		},
	}
	inspector := dockerplatformstesting.NewMockPlatformInspector(ctrl)
	inspector.EXPECT().GetPlatforms(gomock.Any(), "golang").Return(
		dockerplatforms.MustParseDockerPlatformList("linux/amd64, linux/arm/v7, linux/arm64/v8, linux/386, linux/mips64le, linux/ppc64le, linux/s390x, windows/amd64, windows/amd64"),
		nil,
	).AnyTimes()
	inspector.EXPECT().GetPlatforms(gomock.Any(), "golang:1.5").Return(
		dockerplatforms.MustParseDockerPlatformList("linux/amd64"),
		nil,
	).AnyTimes()
	processors := []k8splatforms.KindProcessor{
		&k8splatforms.PodProcessor{},
		&k8splatforms.ReplicaSetProcessor{},
		&k8splatforms.DeploymentProcessor{},
		&k8splatforms.StatefulSetProcessor{},
		&k8splatforms.DaemonSetProcessor{},
		&k8splatforms.JobProcessor{},
		&k8splatforms.CronJobProcessor{},
	}
	testcases := []struct {
		name      string
		objs      []client.Object
		nodes     []corev1.Node
		metricses []metricsv1beta1.PodMetrics
		after     time.Time
		expected  []k8splatforms.Row
	}{
		{
			name:      "empty",
			objs:      nil,
			nodes:     nil,
			metricses: nil,
			after:     time1,
			expected:  nil,
		},
		{
			name: "pod",
			objs: []client.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod1",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time1),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "container1",
								Image: "golang",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod2",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time1),
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
						Containers: []corev1.Container{
							{
								Name:  "container1",
								Image: "golang",
							},
							{
								Name:  "container2",
								Image: "golang:1.5",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
						Labels: map[string]string{
							"kubernetes.io/os":   "linux",
							"kubernetes.io/arch": "amd64",
						},
					},
				},
			},
			metricses: []metricsv1beta1.PodMetrics{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
					Containers: []metricsv1beta1.ContainerMetrics{
						{
							Name:  "container1",
							Usage: corev1.ResourceList{},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "default",
					},
					Containers: []metricsv1beta1.ContainerMetrics{
						{
							Name: "container1",
							Usage: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewScaledQuantity(500, -3),
								corev1.ResourceMemory: *resource.NewScaledQuantity(123456789, 0),
							},
						},
						{
							Name: "container2",
							Usage: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewScaledQuantity(250, -3),
								corev1.ResourceMemory: *resource.NewScaledQuantity(347111248, 0),
							},
						},
					},
				},
			},
			after: time1,
			expected: []k8splatforms.Row{
				{
					Namespace:         "default",
					APIVersion:        "v1",
					Kind:              "Pod",
					Name:              "pod1",
					DeclaredPlatforms: dockerplatforms.MustParseDockerPlatformList("linux/amd64, linux/arm64"),
					ImagePlatforms:    dockerplatforms.MustParseDockerPlatformList("linux/386, linux/amd64, linux/arm, linux/arm64, linux/mips64le, linux/ppc64le, linux/s390x, windows/amd64"),
					ImagePlatformDetails: map[string]dockerplatforms.DockerPlatformList{
						"container1": dockerplatforms.MustParseDockerPlatformList("linux/386, linux/amd64, linux/arm, linux/arm64, linux/mips64le, linux/ppc64le, linux/s390x, windows/amd64"),
					},
					HasViolation: false,
				},
				{
					Namespace:         "default",
					APIVersion:        "v1",
					Kind:              "Pod",
					Name:              "pod2",
					ScheduledPlatform: ptr(dockerplatforms.MustParseDockerPlatform("linux/amd64")),
					DeclaredPlatforms: dockerplatforms.MustParseDockerPlatformList("linux/amd64, linux/arm64"),
					ImagePlatforms:    dockerplatforms.MustParseDockerPlatformList("linux/amd64"),
					ImagePlatformDetails: map[string]dockerplatforms.DockerPlatformList{
						"container1": dockerplatforms.MustParseDockerPlatformList("linux/386, linux/amd64, linux/arm, linux/arm64, linux/mips64le, linux/ppc64le, linux/s390x, windows/amd64"),
						"container2": dockerplatforms.MustParseDockerPlatformList("linux/amd64"),
					},
					HasViolation: true,
					CPUUsage:     0.75,
					MemoryUsage:  470568037.0,
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := k8splatforms.EvaluateObjects(
				ctx,
				tc.objs,
				tc.nodes,
				tc.metricses,
				tc.after,
				nodePlatforms,
				inspector,
				processors,
			)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expected, rows); diff != "" {
				t.Errorf("unexpected rows (-want +got):\n%s", diff)
			}
		})
	}
}

func ptr[T any](value T) *T {
	return &value
}

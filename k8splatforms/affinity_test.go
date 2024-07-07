package k8splatforms_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	"github.com/wantedly/container-platform-tools/k8splatforms"
)

func TestPlatformAffinity_Normalize(t *testing.T) {
	testcases := []struct {
		name     string
		input    k8splatforms.PlatformAffinity
		expected k8splatforms.PlatformAffinity
	}{
		{
			name:     "empty",
			input:    pa(t, ""),
			expected: pa(t, ""),
		},
		{
			name:     "singular single",
			input:    pa(t, "linux/amd64"),
			expected: pa(t, "linux/amd64"),
		},
		{
			name:     "singular multiple",
			input:    pa(t, "linux/amd64, linux/arm64"),
			expected: pa(t, "linux/amd64, linux/arm64"),
		},
		{
			name:     "singular sort",
			input:    pa(t, "windows/amd64, linux/amd64"),
			expected: pa(t, "linux/amd64, windows/amd64"),
		},
		{
			name:     "singular dedupe",
			input:    pa(t, "linux/amd64, linux/amd64"),
			expected: pa(t, "linux/amd64"),
		},
		{
			name:     "arch wildcard single",
			input:    pa(t, "linux/*"),
			expected: pa(t, "linux/*"),
		},
		{
			name:     "arch wildcard multiple",
			input:    pa(t, "linux/*, windows/*"),
			expected: pa(t, "linux/*, windows/*"),
		},
		{
			name:     "arch wildcard dedupe",
			input:    pa(t, "linux/*, linux/*"),
			expected: pa(t, "linux/*"),
		},
		{
			name:     "arch wildcard absorb and sort",
			input:    pa(t, "linux/amd64, windows/amd64, linux/*"),
			expected: pa(t, "linux/*, windows/amd64"),
		},
		{
			name:     "os wildcard single",
			input:    pa(t, "*/amd64"),
			expected: pa(t, "*/amd64"),
		},
		{
			name:     "os wildcard multiple",
			input:    pa(t, "*/amd64, */arm64"),
			expected: pa(t, "*/amd64, */arm64"),
		},
		{
			name:     "os wildcard dedupe",
			input:    pa(t, "*/amd64, */amd64"),
			expected: pa(t, "*/amd64"),
		},
		{
			name:     "os wildcard absorb and sort",
			input:    pa(t, "linux/amd64, linux/arm64, */amd64"),
			expected: pa(t, "*/amd64, linux/arm64"),
		},
		{
			name:     "full wildcard single",
			input:    pa(t, "*/*"),
			expected: pa(t, "*/*"),
		},
		{
			name:     "full wildcard dedupe",
			input:    pa(t, "*/*, */*"),
			expected: pa(t, "*/*"),
		},
		{
			name:     "full wildcard absorb and sort",
			input:    pa(t, "linux/amd64, windows/*, */arm64, */*"),
			expected: pa(t, "*/*"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.Normalize()
			if diff := cmp.Diff(dockerplatforms.DockerPlatformList(tc.expected).String(), dockerplatforms.DockerPlatformList(actual).String()); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlatformAffinity_Union(t *testing.T) {
	testcases := []struct {
		name     string
		a        k8splatforms.PlatformAffinity
		b        k8splatforms.PlatformAffinity
		expected k8splatforms.PlatformAffinity
	}{
		{
			name:     "empty",
			a:        pa(t, ""),
			b:        pa(t, ""),
			expected: pa(t, ""),
		},
		{
			name:     "singular",
			a:        pa(t, "linux/amd64, linux/arm64"),
			b:        pa(t, "linux/amd64, linux/arm"),
			expected: pa(t, "linux/amd64, linux/arm, linux/arm64"),
		},
		{
			name:     "wildcard A",
			a:        pa(t, "*/*"),
			b:        pa(t, "linux/amd64"),
			expected: pa(t, "*/*"),
		},
		{
			name:     "wildcard B",
			a:        pa(t, "linux/amd64"),
			b:        pa(t, "*/*"),
			expected: pa(t, "*/*"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.a.Union(tc.b)
			if diff := cmp.Diff(dockerplatforms.DockerPlatformList(tc.expected).String(), dockerplatforms.DockerPlatformList(actual).String()); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlatformAffinity_Intersect(t *testing.T) {
	testcases := []struct {
		name     string
		a        k8splatforms.PlatformAffinity
		b        k8splatforms.PlatformAffinity
		expected k8splatforms.PlatformAffinity
	}{
		{
			name:     "empty both",
			a:        pa(t, ""),
			b:        pa(t, ""),
			expected: pa(t, ""),
		},
		{
			name:     "empty A",
			a:        pa(t, ""),
			b:        pa(t, "linux/amd64"),
			expected: pa(t, ""),
		},
		{
			name:     "empty B",
			a:        pa(t, "linux/amd64"),
			b:        pa(t, ""),
			expected: pa(t, ""),
		},
		{
			name:     "double wildcard match",
			a:        pa(t, "*/*"),
			b:        pa(t, "*/*"),
			expected: pa(t, "*/*"),
		},
		{
			name:     "mix",
			a:        pa(t, "linux/*"),
			b:        pa(t, "*/amd64"),
			expected: pa(t, "linux/amd64"),
		},
		{
			name:     "mix distrib",
			a:        pa(t, "linux/*, windows/*"),
			b:        pa(t, "*/amd64, */arm64"),
			expected: pa(t, "linux/amd64, linux/arm64, windows/amd64, windows/arm64"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.a.Intersect(tc.b)
			if diff := cmp.Diff(dockerplatforms.DockerPlatformList(tc.expected).String(), dockerplatforms.DockerPlatformList(actual).String()); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func pa(t *testing.T, s string) k8splatforms.PlatformAffinity {
	t.Helper()

	list, err := dockerplatforms.ParseDockerPlatformList(s)
	if err != nil {
		t.Fatal(err)
	}
	return k8splatforms.PlatformAffinity(list)
}

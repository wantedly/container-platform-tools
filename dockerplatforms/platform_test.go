package dockerplatforms_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
)

func TestDockerPlatformString(t *testing.T) {
	testcases := []struct {
		platform dockerplatforms.DockerPlatform
		expected string
	}{
		{
			platform: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
			expected: "linux/amd64",
		},
		{
			platform: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm64",
			},
			expected: "linux/arm64",
		},
		{
			platform: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			expected: "linux/arm/v7",
		},
	}

	for _, tc := range testcases {
		if diff := cmp.Diff(tc.expected, tc.platform.String()); diff != "" {
			t.Errorf("unexpected platform string: %s", diff)
		}
	}
}

func TestParseDockerPlatform(t *testing.T) {
	testcases := []struct {
		platform string
		expected dockerplatforms.DockerPlatform
	}{
		{
			platform: "linux/amd64",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			platform: "linux/arm64",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm64",
			},
		},
		{
			platform: "linux/arm/v7",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
		},
	}

	for _, tc := range testcases {
		p, err := dockerplatforms.ParseDockerPlatform(tc.platform)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if diff := cmp.Diff(tc.expected, p); diff != "" {
			t.Errorf("unexpected platform: %s", diff)
		}
	}
}

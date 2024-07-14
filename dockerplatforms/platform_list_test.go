package dockerplatforms_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	"gopkg.in/yaml.v3"
)

func TestDockerPlatformListString(t *testing.T) {
	testcases := []struct {
		name      string
		platforms dockerplatforms.DockerPlatformList
		string    string
		json      string
		yaml      string
	}{
		{
			name:      "empty",
			platforms: dockerplatforms.DockerPlatformList(nil),
			string:    "",
			json:      `""`,
			yaml:      "\"\"\n",
		},
		{
			name: "single entry",
			platforms: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			string: "linux/amd64",
			json:   `"linux/amd64"`,
			yaml:   "linux/amd64\n",
		},
		{
			name: "multiple entries",
			platforms: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm64",
				},
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
			},
			string: "linux/arm64, linux/arm/v7",
			json:   `"linux/arm64, linux/arm/v7"`,
			yaml:   "linux/arm64, linux/arm/v7\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("String", func(t *testing.T) {
				s := tc.platforms.String()
				if diff := cmp.Diff(tc.string, s); diff != "" {
					t.Errorf("unexpected platform string: %s", diff)
				}
			})

			t.Run("JSON", func(t *testing.T) {
				json, err := json.Marshal(tc.platforms)
				if err != nil {
					t.Error(err)
				}
				if diff := cmp.Diff(tc.json, string(json)); diff != "" {
					t.Errorf("unexpected platform JSON: %s", diff)
				}
			})

			t.Run("YAML", func(t *testing.T) {
				yaml, err := yaml.Marshal(tc.platforms)
				if err != nil {
					t.Error(err)
				}
				if diff := cmp.Diff(tc.yaml, string(yaml)); diff != "" {
					t.Errorf("unexpected platform YAML: %s", diff)
				}
			})
		})
	}
}

func TestParseDockerPlatformList(t *testing.T) {
	testcases := []struct {
		name     string
		string   string
		json     string
		yaml     string
		expected dockerplatforms.DockerPlatformList
	}{
		{
			name:     "String: empty",
			string:   " ",
			expected: dockerplatforms.DockerPlatformList(nil),
		},
		{
			name:   "String: single entry",
			string: "linux/amd64",
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
		},
		{
			name:   "String: multiple entries",
			string: "linux/arm64, linux/arm/v7",
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm64",
				},
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
			},
		},
		{
			name:     "JSON: empty",
			json:     `""`,
			expected: dockerplatforms.DockerPlatformList(nil),
		},
		{
			name: "JSON: single entry",
			json: `"linux/amd64"`,
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
		},
		{
			name: "JSON: multiple entries",
			json: `"linux/arm64, linux/arm/v7"`,
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm64",
				},
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
			},
		},
		{
			name:     "YAML: empty",
			yaml:     `""`,
			expected: dockerplatforms.DockerPlatformList(nil),
		},
		{
			name: "YAML: single entry",
			yaml: "linux/amd64",
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
		},
		{
			name: "YAML: multiple entries",
			yaml: "linux/arm64, linux/arm/v7",
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm64",
				},
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
			},
		},
	}

	for _, tc := range testcases {
		var p dockerplatforms.DockerPlatformList
		if tc.string != "" {
			var err error
			p, err = dockerplatforms.ParseDockerPlatformList(tc.string)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		} else if tc.json != "" {
			if err := json.Unmarshal([]byte(tc.json), &p); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		} else if tc.yaml != "" {
			if err := yaml.Unmarshal([]byte(tc.yaml), &p); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}
		if diff := cmp.Diff(tc.expected, p); diff != "" {
			t.Errorf("unexpected platform: %s", diff)
		}
	}
}

func TestDockerPlatformListNormalized(t *testing.T) {
	testcases := []struct {
		name      string
		platforms dockerplatforms.DockerPlatformList
		expected  dockerplatforms.DockerPlatformList
	}{
		{
			name:      "empty",
			platforms: dockerplatforms.DockerPlatformList(nil),
			expected:  dockerplatforms.DockerPlatformList(nil),
		},
		{
			name: "single entry",
			platforms: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
		},
		{
			name: "multiple entries sorted",
			platforms: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm64",
				},
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
			},
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
				{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
		},
		{
			name: "duplicate entries",
			platforms: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm64",
				},
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
				{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
			expected: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
				{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			normalized := tc.platforms.Normalized()
			if diff := cmp.Diff(tc.expected, normalized); diff != "" {
				t.Errorf("unexpected normalized platform: %s", diff)
			}
		})
	}
}

func TestDockerPlatformListIncludes(t *testing.T) {
	testcases := []struct {
		name     string
		lhs      dockerplatforms.DockerPlatformList
		rhs      dockerplatforms.DockerPlatformList
		expected bool
	}{
		{
			name:     "both empty",
			lhs:      dockerplatforms.DockerPlatformList(nil),
			rhs:      dockerplatforms.DockerPlatformList(nil),
			expected: true,
		},
		{
			name: "lhs empty",
			lhs:  dockerplatforms.DockerPlatformList(nil),
			rhs: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			expected: false,
		},
		{
			name: "rhs empty",
			lhs: dockerplatforms.DockerPlatformList{
				{
					OS:           "linux",
					Architecture: "amd64",
				},
			},
			rhs:      dockerplatforms.DockerPlatformList(nil),
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			included := tc.lhs.Includes(tc.rhs)
			if diff := cmp.Diff(tc.expected, included); diff != "" {
				t.Errorf("unexpected included: %s", diff)
			}
		})
	}
}

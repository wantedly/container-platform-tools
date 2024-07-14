package dockerplatforms_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	"gopkg.in/yaml.v3"
)

func TestDockerPlatformString(t *testing.T) {
	testcases := []struct {
		name     string
		platform dockerplatforms.DockerPlatform
		string   string
		json     string
		yaml     string
	}{
		{
			name: "linux/amd64",
			platform: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
			string: "linux/amd64",
			json:   `"linux/amd64"`,
			yaml:   "linux/amd64\n",
		},
		{
			name: "linux/arm64",
			platform: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm64",
			},
			string: "linux/arm64",
			json:   `"linux/arm64"`,
			yaml:   "linux/arm64\n",
		},
		{
			name: "linux/arm/v7",
			platform: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			string: "linux/arm/v7",
			json:   `"linux/arm/v7"`,
			yaml:   "linux/arm/v7\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("String", func(t *testing.T) {
				s := tc.platform.String()
				if diff := cmp.Diff(tc.string, s); diff != "" {
					t.Errorf("unexpected platform string: %s", diff)
				}
			})

			t.Run("JSON", func(t *testing.T) {
				json, err := json.Marshal(tc.platform)
				if err != nil {
					t.Error(err)
				}
				if diff := cmp.Diff(tc.json, string(json)); diff != "" {
					t.Errorf("unexpected platform JSON: %s", diff)
				}
			})

			t.Run("YAML", func(t *testing.T) {
				yaml, err := yaml.Marshal(tc.platform)
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

func TestParseDockerPlatform(t *testing.T) {
	testcases := []struct {
		name     string
		string   string
		json     string
		yaml     string
		expected dockerplatforms.DockerPlatform
	}{
		{
			name:   "Parse String linux/amd64",
			string: "linux/amd64",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			name:   "Parse String linux/arm64",
			string: "linux/arm64",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm64",
			},
		},
		{
			name:   "Parse String linux/arm/v7",
			string: "linux/arm/v7",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
		},
		{
			name: "Parse JSON linux/amd64",
			json: `"linux/amd64"`,
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			name: "Parse JSON linux/arm64",
			json: `"linux/arm64"`,
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm64",
			},
		},
		{
			name: "Parse JSON linux/arm/v7",
			json: `"linux/arm/v7"`,
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
		},
		{
			name: "Parse YAML linux/amd64",
			yaml: "linux/amd64",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			name: "Parse YAML linux/arm64",
			yaml: "linux/arm64",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm64",
			},
		},
		{
			name: "Parse YAML linux/arm/v7",
			yaml: "linux/arm/v7",
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
		},
	}

	for _, tc := range testcases {
		var p dockerplatforms.DockerPlatform
		if tc.string != "" {
			var err error
			p, err = dockerplatforms.ParseDockerPlatform(tc.string)
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

func TestDockerPlatformVariantless(t *testing.T) {
	testcases := []struct {
		name     string
		input    dockerplatforms.DockerPlatform
		expected dockerplatforms.DockerPlatform
	}{
		{
			name: "linux/amd64",
			input: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			name: "linux/arm/v7",
			input: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			expected: dockerplatforms.DockerPlatform{
				OS:           "linux",
				Architecture: "arm",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.input.Variantless()
			if diff := cmp.Diff(tc.expected, p); diff != "" {
				t.Errorf("unexpected platform: %s", diff)
			}
		})
	}
}

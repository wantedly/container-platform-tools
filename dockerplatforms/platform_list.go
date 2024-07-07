package dockerplatforms

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

type DockerPlatformList []DockerPlatform

var _ json.Unmarshaler = &DockerPlatformList{}
var _ json.Marshaler = &DockerPlatformList{}
var _ yaml.Unmarshaler = &DockerPlatformList{}
var _ yaml.Marshaler = &DockerPlatformList{}

func (l DockerPlatformList) String() string {
	platforms := make([]string, len(l))
	for i, platform := range l {
		platforms[i] = platform.String()
	}
	return strings.Join(platforms, ", ")
}

func ParseDockerPlatformList(s string) (DockerPlatformList, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	platformTexts := strings.Split(s, ",")
	platforms := make([]DockerPlatform, len(platformTexts))
	for i, platformText := range platformTexts {
		platformText = strings.TrimSpace(platformText)
		platform, err := ParseDockerPlatform(platformText)
		if err != nil {
			return nil, err
		}
		platforms[i] = platform
	}
	return platforms, nil
}

func (p DockerPlatformList) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *DockerPlatformList) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	parsed, err := ParseDockerPlatformList(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

func (p DockerPlatformList) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

func (p *DockerPlatformList) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}

	parsed, err := ParseDockerPlatformList(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

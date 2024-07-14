package dockerplatforms

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type DockerPlatform struct {
	// OS specifies the operating system,
	// using the naming convention from Go's GOOS.
	// Currently known values are:
	// - android
	// - darwin
	// - dragonfly
	// - freebsd
	// - illumos
	// - ios
	// - js
	// - linux
	// - netbsd
	// - openbsd
	// - plan9
	// - solaris
	// - wasip1
	// - windows
	OS string
	// Architecture specifies the CPU architecture,
	// using the naming convention from Go's GOARCH.
	// Currently known values are:
	// - 386
	// - amd64
	// - arm
	// - arm64
	// - mips
	// - mips64
	// - mips64le
	// - mipsle
	// - ppc64
	// - ppc64le
	// - s390x
	// - wasm
	Architecture string
	// Variant specifies variant of the CPU architecture.
	// For the possible values, see the OCI image spec:
	// https://github.com/opencontainers/image-spec/blob/v1.0.0/image-index.md
	// When not used, the value is empty.
	//
	// Currently known combinations are:
	// - */arm/v6
	// - */arm/v7
	// - */arm/v8
	// - */arm64/v8
	Variant string
}

var _ json.Unmarshaler = &DockerPlatform{}
var _ json.Marshaler = &DockerPlatform{}
var _ yaml.Unmarshaler = &DockerPlatform{}
var _ yaml.Marshaler = &DockerPlatform{}

func (p DockerPlatform) String() string {
	if p.Variant != "" {
		return fmt.Sprintf("%s/%s/%s", p.OS, p.Architecture, p.Variant)
	}
	return fmt.Sprintf("%s/%s", p.OS, p.Architecture)
}

func ParseDockerPlatform(platform string) (DockerPlatform, error) {
	// split by slash
	parts := strings.Split(platform, "/")
	if len(parts) < 2 {
		return DockerPlatform{}, fmt.Errorf("invalid platform: too few parts: %s", platform)
	}
	if len(parts) > 3 {
		return DockerPlatform{}, fmt.Errorf("invalid platform: too many parts: %s", platform)
	}
	result := DockerPlatform{
		OS:           parts[0],
		Architecture: parts[1],
	}
	if len(parts) >= 3 {
		result.Variant = parts[2]
	}
	return result, nil
}

func (p DockerPlatform) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *DockerPlatform) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	parsed, err := ParseDockerPlatform(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

func (p DockerPlatform) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

func (p *DockerPlatform) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}

	parsed, err := ParseDockerPlatform(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

func (p DockerPlatform) Cmp(other DockerPlatform) int {
	if p.OS < other.OS {
		return -1
	} else if p.OS > other.OS {
		return 1
	}
	if p.Architecture < other.Architecture {
		return -1
	} else if p.Architecture > other.Architecture {
		return 1
	}
	if p.Variant < other.Variant {
		return -1
	} else if p.Variant > other.Variant {
		return 1
	}
	return 0
}

func (p DockerPlatform) Variantless() DockerPlatform {
	p.Variant = ""
	return p
}

func MustParseDockerPlatform(platform string) DockerPlatform {
	p, err := ParseDockerPlatform(platform)
	if err != nil {
		panic(err)
	}
	return p
}

var (
	Linux386      = MustParseDockerPlatform("linux/386")
	LinuxAMD64    = MustParseDockerPlatform("linux/amd64")
	LinuxAMD64V2  = MustParseDockerPlatform("linux/amd64/v2")
	LinuxAMD64V3  = MustParseDockerPlatform("linux/amd64/v3")
	LinuxAMD64V4  = MustParseDockerPlatform("linux/amd64/v4")
	LinuxARM64    = MustParseDockerPlatform("linux/arm64")
	LinuxARM64V8  = MustParseDockerPlatform("linux/arm64/v8")
	LinuxARMV5    = MustParseDockerPlatform("linux/arm/v5")
	LinuxARMV6    = MustParseDockerPlatform("linux/arm/v6")
	LinuxARMV7    = MustParseDockerPlatform("linux/arm/v7")
	LinuxMIPS     = MustParseDockerPlatform("linux/mips")
	LinuxMIPS64   = MustParseDockerPlatform("linux/mips64")
	LinuxMIPS64LE = MustParseDockerPlatform("linux/mips64le")
	LinuxMIPSLE   = MustParseDockerPlatform("linux/mipsle")
	LinuxPPC64LE  = MustParseDockerPlatform("linux/ppc64le")
	LinuxRISCV64  = MustParseDockerPlatform("linux/riscv64")
	LinuxS390X    = MustParseDockerPlatform("linux/s390x")
	WindowsAMD64  = MustParseDockerPlatform("windows/amd64")
)

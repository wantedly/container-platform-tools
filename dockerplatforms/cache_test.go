package dockerplatforms_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
)

func TestNopGetEmpty(t *testing.T) {
	cache := dockerplatforms.NewNopCache()

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ok, false); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform(nil)); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}
}

func TestNopSet(t *testing.T) {
	cache := dockerplatforms.NewNopCache()

	err := cache.SetCachedPlatforms("docker.io/library/golang:latest", []dockerplatforms.DockerPlatform{})
	if err != nil {
		t.Fatal(err)
	}

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ok, false); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform(nil)); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}
}

func TestNopSetError(t *testing.T) {
	cache := dockerplatforms.NewNopCache()

	cache.SetErrorCache("docker.io/library/golang:latest", errors.New("failed to retrieve"))

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ok, false); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform(nil)); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}
}

func TestNopClear(t *testing.T) {
	cache := dockerplatforms.NewNopCache()

	err := cache.ClearCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ok, false); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform(nil)); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}
}

func TestYAMLGetEmpty(t *testing.T) {
	env, cache, err := setupEmptyYAML("TestYAMLGetEmpty")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ok, false); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform(nil)); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}
}

func TestYAMLEmptyWrite(t *testing.T) {
	env, _, err := setupEmptyYAML("TestYAMLEmptyWrite")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	content, err := os.ReadFile(env.path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(content), ""); diff != "" {
		t.Errorf("cache file (-want +got):\n%s", diff)
	}
}

func TestYAMLSetGet(t *testing.T) {
	env, cache, err := setupEmptyYAML("TestYAMLSetGet")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	err = cache.SetCachedPlatforms("docker.io/library/golang:latest", []dockerplatforms.DockerPlatform{
		dockerplatforms.Linux386,
		dockerplatforms.LinuxAMD64,
		dockerplatforms.LinuxARM64,
		dockerplatforms.LinuxARMV7,
		dockerplatforms.LinuxMIPS64LE,
		dockerplatforms.LinuxPPC64LE,
		dockerplatforms.LinuxS390X,
		dockerplatforms.WindowsAMD64,
	})
	if err != nil {
		t.Fatal(err)
	}

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ok, true); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform{
		dockerplatforms.Linux386,
		dockerplatforms.LinuxAMD64,
		dockerplatforms.LinuxARM64,
		dockerplatforms.LinuxARMV7,
		dockerplatforms.LinuxMIPS64LE,
		dockerplatforms.LinuxPPC64LE,
		dockerplatforms.LinuxS390X,
		dockerplatforms.WindowsAMD64,
	}); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}
}

func TestYAMLSetWrite(t *testing.T) {
	env, cache, err := setupEmptyYAML("TestYAMLSetWrite")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	err = cache.SetCachedPlatforms("docker.io/library/golang:latest", []dockerplatforms.DockerPlatform{
		dockerplatforms.Linux386,
		dockerplatforms.LinuxAMD64,
		dockerplatforms.LinuxARM64,
		dockerplatforms.LinuxARMV7,
		dockerplatforms.LinuxMIPS64LE,
		dockerplatforms.LinuxPPC64LE,
		dockerplatforms.LinuxS390X,
		dockerplatforms.WindowsAMD64,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cache.WriteBack()
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(env.path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(content), `docker.io/library/golang:latest:
    platforms: linux/386, linux/amd64, linux/arm64, linux/arm/v7, linux/mips64le, linux/ppc64le, linux/s390x, windows/amd64
`); diff != "" {
		t.Errorf("cache file (-want +got):\n%s", diff)
	}
}

func TestYAMLReadGet(t *testing.T) {
	env, cache, err := setupYAML("TestYAMLReadGet", `docker.io/library/golang:latest:
    platforms: linux/386, linux/amd64, linux/arm64, linux/arm/v7, linux/mips64le, linux/ppc64le, linux/s390x, windows/amd64
`)
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ok, true); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform{
		dockerplatforms.Linux386,
		dockerplatforms.LinuxAMD64,
		dockerplatforms.LinuxARM64,
		dockerplatforms.LinuxARMV7,
		dockerplatforms.LinuxMIPS64LE,
		dockerplatforms.LinuxPPC64LE,
		dockerplatforms.LinuxS390X,
		dockerplatforms.WindowsAMD64,
	}); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}
}

func TestYAMLSetErrorWrite(t *testing.T) {
	env, cache, err := setupEmptyYAML("TestYAMLSetErrorWrite")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	cache.SetErrorCache("docker.io/library/golang:latest", errors.New("failed to retrieve"))

	err = cache.WriteBack()
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(env.path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(content), `docker.io/library/golang:latest:
    platforms: ""
    error: failed to retrieve
`); diff != "" {
		t.Errorf("cache file (-want +got):\n%s", diff)
	}
}

func TestYAMLReadErrorGet(t *testing.T) {
	env, cache, err := setupYAML("TestYAMLReadErrorGet", `docker.io/library/golang:latest:
    platforms: ""
    error: failed to retrieve
`)
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	_, _, err = cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if err == nil {
		t.Fatal("Missing error")
	}
	if diff := cmp.Diff(err.Error(), "failed to retrieve"); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
}

type yamlEnv struct {
	tmpdir string
	path   string
}

func (e *yamlEnv) Close() error {
	return os.RemoveAll(e.tmpdir)
}

func setupEmptyYAML(name string) (*yamlEnv, *dockerplatforms.YAMLCache, error) {
	return setupYAMLImpl(name, false, "")
}

func setupYAML(name string, content string) (*yamlEnv, *dockerplatforms.YAMLCache, error) {
	return setupYAMLImpl(name, true, content)
}
func setupYAMLImpl(name string, hasContent bool, content string) (*yamlEnv, *dockerplatforms.YAMLCache, error) {
	tmpdir, err := os.MkdirTemp("", name)
	if err != nil {
		return nil, nil, err
	}
	path := tmpdir + "/cache.yaml"

	if hasContent {
		err = os.WriteFile(path, []byte(content), 0o666)
		if err != nil {
			os.RemoveAll(tmpdir)
			return nil, nil, err
		}
	}

	initCache, err := dockerplatforms.NewYAMLCache(path)
	if err != nil {
		os.RemoveAll(tmpdir)
		return nil, nil, err
	}
	return &yamlEnv{tmpdir, path}, initCache, nil
}

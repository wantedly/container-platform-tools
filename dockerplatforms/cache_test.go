package dockerplatforms_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
)

func TestNopGetEmpty(t *testing.T) {
	ctx := context.Background()
	cache := dockerplatforms.NewNopCache()

	platforms, ok, err := cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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
	ctx := context.Background()
	cache := dockerplatforms.NewNopCache()

	err := cache.SetCachedPlatforms(ctx, "docker.io/library/golang:latest", []dockerplatforms.DockerPlatform{})
	if err != nil {
		t.Fatal(err)
	}

	platforms, ok, err := cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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
	ctx := context.Background()
	cache := dockerplatforms.NewNopCache()

	cache.SetErrorCache(ctx, "docker.io/library/golang:latest", errors.New("failed to retrieve"))

	platforms, ok, err := cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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
	ctx := context.Background()
	cache := dockerplatforms.NewNopCache()

	err := cache.ClearCachedPlatforms(ctx, "docker.io/library/golang:latest")
	if err != nil {
		t.Fatal(err)
	}

	platforms, ok, err := cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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
	ctx := context.Background()
	env, cache, err := setupEmptyYAML(ctx, "TestYAMLGetEmpty")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	platforms, ok, err := cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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
	ctx := context.Background()
	env, _, err := setupEmptyYAML(ctx, "TestYAMLEmptyWrite")
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
	ctx := context.Background()
	env, cache, err := setupEmptyYAML(ctx, "TestYAMLSetGet")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	err = cache.SetCachedPlatforms(ctx, "docker.io/library/golang:latest", []dockerplatforms.DockerPlatform{
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

	platforms, ok, err := cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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
	ctx := context.Background()
	env, cache, err := setupEmptyYAML(ctx, "TestYAMLSetWrite")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	err = cache.SetCachedPlatforms(ctx, "docker.io/library/golang:latest", []dockerplatforms.DockerPlatform{
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

	err = cache.WriteBack(ctx)
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
	ctx := context.Background()
	env, cache, err := setupYAML(ctx, "TestYAMLReadGet", `docker.io/library/golang:latest:
    platforms: linux/386, linux/amd64, linux/arm64, linux/arm/v7, linux/mips64le, linux/ppc64le, linux/s390x, windows/amd64
`)
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	platforms, ok, err := cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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
	ctx := context.Background()
	env, cache, err := setupEmptyYAML(ctx, "TestYAMLSetErrorWrite")
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	cache.SetErrorCache(ctx, "docker.io/library/golang:latest", errors.New("failed to retrieve"))

	err = cache.WriteBack(ctx)
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
	ctx := context.Background()
	env, cache, err := setupYAML(ctx, "TestYAMLReadErrorGet", `docker.io/library/golang:latest:
    platforms: ""
    error: failed to retrieve
`)
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()

	_, _, err = cache.GetCachedPlatforms(ctx, "docker.io/library/golang:latest")
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

func setupEmptyYAML(ctx context.Context, name string) (*yamlEnv, *dockerplatforms.YAMLCache, error) {
	return setupYAMLImpl(ctx, name, false, "")
}

func setupYAML(ctx context.Context, name string, content string) (*yamlEnv, *dockerplatforms.YAMLCache, error) {
	return setupYAMLImpl(ctx, name, true, content)
}
func setupYAMLImpl(ctx context.Context, name string, hasContent bool, content string) (*yamlEnv, *dockerplatforms.YAMLCache, error) {
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

	initCache, err := dockerplatforms.NewYAMLCache(ctx, path)
	if err != nil {
		os.RemoveAll(tmpdir)
		return nil, nil, err
	}
	return &yamlEnv{tmpdir, path}, initCache, nil
}

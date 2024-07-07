package dockerplatforms_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
)

func TestNopCache(t *testing.T) {
	cache := dockerplatforms.NewNopCache()

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if diff := cmp.Diff(err, nil); diff != "" {
		t.Errorf("GetCachedPlatforms() error (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(ok, false); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform(nil)); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}

	err = cache.SetCachedPlatforms("docker.io/library/golang:latest", []dockerplatforms.DockerPlatform{})
	if diff := cmp.Diff(err, nil); diff != "" {
		t.Errorf("SetCachedPlatforms() error (-want +got):\n%s", diff)
	}
}

func TestYAMLCache(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "TestYAMLCache")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	cache, err := dockerplatforms.NewYAMLCache(tmpdir + "/cache.yaml")
	if err != nil {
		t.Fatal(err)
	}

	platforms, ok, err := cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if diff := cmp.Diff(err, nil); diff != "" {
		t.Errorf("GetCachedPlatforms() error (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(ok, false); diff != "" {
		t.Errorf("GetCachedPlatforms() ok (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform(nil)); diff != "" {
		t.Errorf("GetCachedPlatforms() platforms (-want +got):\n%s", diff)
	}

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
	if diff := cmp.Diff(err, nil); diff != "" {
		t.Errorf("SetCachedPlatforms() error (-want +got):\n%s", diff)
	}

	platforms, ok, err = cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if diff := cmp.Diff(err, nil); diff != "" {
		t.Errorf("GetCachedPlatforms() error (-want +got):\n%s", diff)
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

	err = cache.WriteBack()
	if diff := cmp.Diff(err, nil); diff != "" {
		t.Errorf("WriteBack() error (-want +got):\n%s", diff)
	}

	cache, err = dockerplatforms.NewYAMLCache(tmpdir + "/cache.yaml")
	if err != nil {
		t.Fatal(err)
	}

	platforms, ok, err = cache.GetCachedPlatforms("docker.io/library/golang:latest")
	if diff := cmp.Diff(err, nil); diff != "" {
		t.Errorf("GetCachedPlatforms() error (-want +got):\n%s", diff)
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

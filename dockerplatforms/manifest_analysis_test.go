package dockerplatforms_test

import (
	"context"
	"os"
	"testing"

	"github.com/containers/image/v5/docker/reference"
	"github.com/google/go-cmp/cmp"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	dockerplatformstesting "github.com/wantedly/container-platform-tools/dockerplatforms/testing"
	"go.uber.org/mock/gomock"
)

func TestAnalyzeManifestIndex(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	retriever := dockerplatformstesting.NewMockManifestRetriever(ctrl)
	retriever.EXPECT().GetManifest(gomock.Any(), "docker.io/library/golang:latest").Return(mustReadFixture(t, "golang-latest.json"), "", nil)
	imageRef, err := reference.ParseNormalizedNamed("golang:latest")
	if err != nil {
		t.Fatal(err)
	}

	platforms, err := dockerplatforms.AnalyzeManifest(ctx, imageRef, retriever)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform{
		dockerplatforms.LinuxAMD64,
		dockerplatforms.LinuxARMV7,
		dockerplatforms.LinuxARM64V8,
		dockerplatforms.Linux386,
		dockerplatforms.LinuxMIPS64LE,
		dockerplatforms.LinuxPPC64LE,
		dockerplatforms.LinuxS390X,
		dockerplatforms.WindowsAMD64,
		dockerplatforms.WindowsAMD64,
	}); diff != "" {
		t.Errorf("AnalyzeManifest() (-want +got):\n%s", diff)
	}
}

func TestAnalyzeManifestV2(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	retriever := dockerplatformstesting.NewMockManifestRetriever(ctrl)
	retriever.EXPECT().GetManifest(gomock.Any(), "docker.io/library/golang:1.5").Return(mustReadFixture(t, "golang-1.5.json"), "", nil)
	retriever.EXPECT().GetManifest(gomock.Any(), "docker.io/library/golang@sha256:99668503de157252ba311f570f036490602095f2620c46cb407d3d2dd88aeb6e").Return(mustReadFixture(t, "golang-1.5-99668503de157252ba311f570f036490602095f2620c46cb407d3d2dd88aeb6e.json"), "", nil)
	imageRef, err := reference.ParseNormalizedNamed("golang:1.5")
	if err != nil {
		t.Fatal(err)
	}

	platforms, err := dockerplatforms.AnalyzeManifest(ctx, imageRef, retriever)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform{
		dockerplatforms.LinuxAMD64,
	}); diff != "" {
		t.Errorf("AnalyzeManifest() (-want +got):\n%s", diff)
	}
}

func TestAnalyzeManifestV1(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	retriever := dockerplatformstesting.NewMockManifestRetriever(ctrl)
	retriever.EXPECT().GetManifest(ctx, "docker.io/library/golang:1.3").Return(mustReadFixture(t, "golang-1.3.json"), "", nil)
	imageRef, err := reference.ParseNormalizedNamed("golang:1.3")
	if err != nil {
		t.Fatal(err)
	}

	platforms, err := dockerplatforms.AnalyzeManifest(ctx, imageRef, retriever)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(platforms, []dockerplatforms.DockerPlatform{
		dockerplatforms.LinuxAMD64,
	}); diff != "" {
		t.Errorf("AnalyzeManifest() (-want +got):\n%s", diff)
	}
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile("fixtures/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

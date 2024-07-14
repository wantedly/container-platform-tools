//go:generate go run go.uber.org/mock/mockgen -source=manifest_retriever.go -destination=testing/mock_manifest_retriever.go -package=dockerplatformstesting
package dockerplatforms

import (
	"context"
	"os/exec"

	"github.com/pkg/errors"
)

// ManifestRetriever is an interface to retrieve the manifest of an image.
type ManifestRetriever interface {
	GetManifest(ctx context.Context, image string) ([]byte, string, error)
}

// ImageTools is a ManifestRetriever backend using `docker buildx imagetools`.
type ImageTools struct {
	dockerExecPath string
}

var _ ManifestRetriever = &ImageTools{}

func NewImageTools() *ImageTools {
	return &ImageTools{dockerExecPath: "docker"}
}

func (i *ImageTools) GetManifest(ctx context.Context, image string) ([]byte, string, error) {
	// Invoke docker buildx imagetools inspect --raw <image>
	cmd := exec.CommandContext(ctx, i.dockerExecPath, "buildx", "imagetools", "inspect", "--raw", image)
	if cmd.Err != nil {
		return nil, "", errors.Wrap(cmd.Err, "creating command")
	}
	result, err := cmd.Output()
	if err != nil {
		return nil, "", errors.Wrap(err, "running imagetools")
	}
	return result, "", nil
}

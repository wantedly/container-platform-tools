package dockerplatforms

import (
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/pkg/errors"
)

func AnalyzeManifest(imageRef reference.Named, retriever ManifestRetriever) ([]DockerPlatform, error) {
	manifestText, mediaType, err := retriever.GetManifest(imageRef.String())
	if err != nil {
		return nil, errors.Wrap(err, "retrieving manifest")
	}
	if mediaType == "" {
		mediaType = manifest.GuessMIMEType(manifestText)
	}

	mediaType = manifest.NormalizedMIMEType(mediaType)

	if manifest.MIMETypeIsMultiImage(mediaType) {
		platforms := []DockerPlatform{}
		list, err := manifest.ListFromBlob(manifestText, mediaType)
		if err != nil {
			return nil, errors.Wrap(err, "parsing manifest list")
		}
		for _, instanceDigest := range list.Instances() {
			instance, err := list.Instance(instanceDigest)
			if err != nil {
				return nil, errors.Wrap(err, "getting instance")
			}
			if instance.ReadOnly.Platform.OS == "unknown" && instance.ReadOnly.Platform.Architecture == "unknown" {
				continue
			}
			platforms = append(platforms, DockerPlatform{
				OS:           instance.ReadOnly.Platform.OS,
				Architecture: instance.ReadOnly.Platform.Architecture,
				Variant:      instance.ReadOnly.Platform.Variant,
			})
		}
		return platforms, nil
	} else {
		manifest, err := manifest.FromBlob(manifestText, mediaType)
		if err != nil {
			return nil, errors.Wrap(err, "parsing manifest")
		}

		image, err := manifest.Inspect(func(blobInfo types.BlobInfo) ([]byte, error) {
			imageNameOnly, err := reference.WithName(imageRef.Name())
			if err != nil {
				return nil, errors.Wrap(err, "getting image name")
			}
			configReference, err := reference.WithDigest(imageNameOnly, manifest.ConfigInfo().Digest)
			if err != nil {
				return nil, errors.Wrap(err, "getting config reference")
			}

			configText, _, err := retriever.GetManifest(configReference.String())
			if err != nil {
				return nil, errors.Wrap(err, "retrieving config")
			}
			return configText, nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "inspecting manifest")
		}
		return []DockerPlatform{
			{
				OS:           image.Os,
				Architecture: image.Architecture,
				Variant:      image.Variant,
			},
		}, nil
	}
}

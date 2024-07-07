package dockerplatforms

import (
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/pkg/errors"
)

type PlatformInspector struct {
	retriever ManifestRetriever
	cache     Cache
}

// New creates a new PlatformInspector.
func New(retriever ManifestRetriever, cache Cache) *PlatformInspector {
	return &PlatformInspector{
		retriever: retriever,
		cache:     cache,
	}
}

// GetPlatforms returns the list of supported platforms of the given image.
func (p *PlatformInspector) GetPlatforms(image string) ([]DockerPlatform, error) {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return nil, errors.Wrap(err, "parsing image reference")
	}
	ref = reference.TagNameOnly(ref)
	// Remove tag if there is a digest
	if ref2, ok := ref.(namedTaggedDigested); ok {
		named, err := reference.WithName(ref2.Name())
		if err != nil {
			return nil, errors.Wrap(err, "removing digest from reference")
		}
		ref, err = reference.WithDigest(named, ref2.Digest())
		if err != nil {
			return nil, errors.Wrap(err, "removing digest from reference")
		}
	}

	image = ref.String()

	cached, ok, err := p.cache.GetCachedPlatforms(image)
	if err != nil {
		return nil, errors.Wrap(err, "looking for cached data")
	}
	if ok {
		return cached, nil
	}

	platforms, err := p.getPlatformsNoCache(image)
	if err != nil {
		p.cache.SetErrorCache(image, err)
		return nil, err
	}
	err = p.cache.SetCachedPlatforms(image, platforms)
	if err != nil {
		return nil, errors.Wrap(err, "caching platforms")
	}

	return platforms, nil
}

func (p *PlatformInspector) getPlatformsNoCache(image string) ([]DockerPlatform, error) {
	manifestText, mediaType, err := p.retriever.GetManifest(image)
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
			imageParsed, err := reference.ParseNamed(image)
			if err != nil {
				return nil, errors.Wrap(err, "parsing image reference")
			}
			imageNameOnly, err := reference.WithName(imageParsed.Name())
			if err != nil {
				return nil, errors.Wrap(err, "getting image name")
			}
			configReference, err := reference.WithDigest(imageNameOnly, manifest.ConfigInfo().Digest)
			if err != nil {
				return nil, errors.Wrap(err, "getting config reference")
			}

			configText, _, err := p.retriever.GetManifest(configReference.String())
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

type namedTaggedDigested interface {
	reference.NamedTagged
	reference.Digested
}

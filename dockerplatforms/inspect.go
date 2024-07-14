package dockerplatforms

import (
	"context"

	"github.com/containers/image/v5/docker/reference"
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
func (p *PlatformInspector) GetPlatforms(ctx context.Context, image string) ([]DockerPlatform, error) {
	imageRef, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return nil, errors.Wrap(err, "parsing image reference")
	}
	imageRef = reference.TagNameOnly(imageRef)
	// Remove tag if there is a digest
	if ref2, ok := imageRef.(namedTaggedDigested); ok {
		named, err := reference.WithName(ref2.Name())
		if err != nil {
			return nil, errors.Wrap(err, "removing digest from reference")
		}
		imageRef, err = reference.WithDigest(named, ref2.Digest())
		if err != nil {
			return nil, errors.Wrap(err, "removing digest from reference")
		}
	}

	image = imageRef.String()

	cached, ok, err := p.cache.GetCachedPlatforms(ctx, image)
	if err != nil {
		return nil, errors.Wrap(err, "looking for cached data")
	}
	if ok {
		return cached, nil
	}

	platforms, err := AnalyzeManifest(ctx, imageRef, p.retriever)
	if err != nil {
		p.cache.SetErrorCache(ctx, image, err)
		return nil, err
	}
	err = p.cache.SetCachedPlatforms(ctx, image, platforms)
	if err != nil {
		return nil, errors.Wrap(err, "caching platforms")
	}

	return platforms, nil
}

type namedTaggedDigested interface {
	reference.NamedTagged
	reference.Digested
}

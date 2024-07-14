package dockerplatforms

import (
	"context"
	"io"
	"os"
	"syscall"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Cache interface {
	// GetCachedPlatforms returns the precomputed list of platforms, if there is any.
	GetCachedPlatforms(ctx context.Context, image string) ([]DockerPlatform, bool, error)
	// SetCachedPlatforms stores the list of platforms for future use.
	SetCachedPlatforms(ctx context.Context, image string, platforms []DockerPlatform) error
	// SetErrorCache stores an error message for the image.
	SetErrorCache(ctx context.Context, image string, err error)
	// ClearCachedPlatforms removes the list of platforms from the cache.
	ClearCachedPlatforms(ctx context.Context, image string) error
}

// NewNopCache creates a Cache that does nothing.
func NewNopCache() Cache {
	return &nopCache{}
}

type nopCache struct{}

func (c *nopCache) GetCachedPlatforms(ctx context.Context, image string) ([]DockerPlatform, bool, error) {
	return nil, false, nil
}

func (c *nopCache) SetCachedPlatforms(ctx context.Context, image string, platforms []DockerPlatform) error {
	return nil
}

func (c *nopCache) ClearCachedPlatforms(ctx context.Context, image string) error {
	return nil
}

func (c *nopCache) SetErrorCache(ctx context.Context, image string, err error) {
}

// NewYAMLCache creates a Cache that reads and writes a YAML file.
func NewYAMLCache(ctx context.Context, path string) (*YAMLCache, error) {
	cache := YAMLCache{
		path:    path,
		oldData: make(map[string]imageData),
		newData: make(map[string]imageData),
	}
	err := cache.open(ctx)
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

var _ Cache = &YAMLCache{}

type YAMLCache struct {
	path    string
	oldData map[string]imageData
	newData map[string]imageData
}

func (c *YAMLCache) open(ctx context.Context) error {
	f, err := os.OpenFile(c.path, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		return errors.Wrap(err, "opening cache file for reading")
	}
	defer f.Close()

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
	if err != nil {
		return errors.Wrap(err, "locking the cache file for reading")
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	yamlText, err := io.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "reading the cache file")
	}
	err = yaml.Unmarshal(yamlText, &c.oldData)
	if err != nil {
		return errors.Wrap(err, "parsing the cache YAML")
	}

	return nil
}
func (c *YAMLCache) WriteBack(ctx context.Context) error {
	if len(c.newData) == 0 {
		return nil
	}

	f, err := os.OpenFile(c.path, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return errors.Wrap(err, "opening cache file for writing")
	}
	defer f.Close()

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	if err != nil {
		return errors.Wrap(err, "locking the cache file for writing")
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	oldYAMLText, err := io.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "reading the cache file")
	}
	err = yaml.Unmarshal(oldYAMLText, &c.oldData)
	if err != nil {
		return errors.Wrap(err, "parsing the cache YAML")
	}

	for image, platforms := range c.newData {
		c.oldData[image] = platforms
	}
	c.newData = make(map[string]imageData)

	newYAMLText, err := yaml.Marshal(c.oldData)
	if err != nil {
		return errors.Wrap(err, "marshalling the cache YAML")
	}
	// Seek the file, write, and truncate to the new size
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "seeking the cache file")
	}
	_, err = f.Write(newYAMLText)
	if err != nil {
		return errors.Wrap(err, "writing the cache file")
	}
	err = f.Truncate(int64(len(newYAMLText)))
	if err != nil {
		return errors.Wrap(err, "truncating the cache file")
	}

	return nil
}

func (c *YAMLCache) GetCachedPlatforms(ctx context.Context, image string) ([]DockerPlatform, bool, error) {
	imageData, ok := c.newData[image]
	if !ok {
		imageData, ok = c.oldData[image]
	}
	if !ok {
		return nil, false, nil
	}

	if imageData.Error != "" {
		return nil, false, errors.New(imageData.Error)
	}

	return imageData.Platforms, true, nil
}

func (c *YAMLCache) SetCachedPlatforms(ctx context.Context, image string, platforms []DockerPlatform) error {
	c.newData[image] = imageData{
		Platforms: platforms,
	}
	return nil
}

// ClearCachedPlatforms implements Cache.
func (c *YAMLCache) ClearCachedPlatforms(ctx context.Context, image string) error {
	delete(c.newData, image)
	return nil
}

// SetErrorCache implements Cache.
func (c *YAMLCache) SetErrorCache(ctx context.Context, image string, err error) {
	c.newData[image] = imageData{
		Error: err.Error(),
	}
}

type imageData struct {
	Platforms DockerPlatformList `json:"platforms" yaml:"platforms"`
	Error     string             `json:"error,omitempty" yaml:"error,omitempty"`
}

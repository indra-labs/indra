package docker

import (
	"github.com/docker/docker/api/types"
)

type BuildConfiguration struct {
	Name            string
	ContextFilePath string
	BuildOpts       types.ImageBuildOptions
	PushOpts        types.ImagePushOptions
}

// FixTagPrefix returns the full set of tags in the BuildConfiguration.
func (bc *BuildConfiguration) FixTagPrefix() (fullTags []string) {

	for _, tag := range bc.BuildOpts.Tags {

		fullTags = append(fullTags, bc.Name+":"+tag)
	}

	return fullTags
}

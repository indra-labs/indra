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

func (self *BuildConfiguration) FixTagPrefix() []string {

	var fullTags = []string{}

	for _, tag := range self.BuildOpts.Tags {

		fullTags = append(fullTags, self.Name+":"+tag)
	}

	return fullTags
}

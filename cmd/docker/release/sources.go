package main

import (
	"github.com/docker/docker/api/types"
	"github.com/indra-labs/indra/pkg/docker"
)

var sourceConfigurations = []docker.BuildConfiguration{
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "btcd-source",
		ContextFilePath: "/tmp/btcd-source.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/btcd/source/official.Dockerfile",
			Tags: []string{
				"v0.23.3",
			},
			BuildArgs: map[string]*string{
				"sourcing_image": strPtr(defaultBuildContainer),
				"source_url":     strPtr("https://github.com/btcsuite/btcd/releases/download"),
				"source_version": strPtr("v0.23.3"),
			},
			SuppressOutput: false,
			Remove:         false,
			ForceRemove:    false,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lnd-source",
		ContextFilePath: "/tmp/lnd-source.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/lnd/source/official.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
			},
			BuildArgs: map[string]*string{
				"sourcing_image": strPtr(defaultBuildContainer),
				"source_url":     strPtr("https://github.com/lightningnetwork/lnd/releases/download"),
				"source_version": strPtr("v0.15.5-beta"),
			},
			SuppressOutput: false,
			Remove:         false,
			ForceRemove:    false,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "indra-source-local",
		ContextFilePath: "/tmp/indra-source-local.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/indra/source/local.Dockerfile",
			Tags: []string{
				"dev",
			},
			BuildArgs: map[string]*string{
				"sourcing_image": strPtr(defaultBuildContainer),
			},
			SuppressOutput: false,
			Remove:         false,
			ForceRemove:    false,
			PullParent:     false,
		},
	},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "indra-source",
	//	ContextFilePath: "/tmp/indra-source.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/indra/source/official.Dockerfile",
	//		Tags: []string{
	//			"v0.1.9",
	//		},
	//		BuildArgs: map[string]*string{
	//			"sourcing_image":            strPtr(defaultBuildContainer),
	//			"source_release_url_prefix": strPtr("https://github.com/indra-labs/indra/releases/download"),
	//			"source_version":            strPtr("v0.1.9"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         false,
	//		ForceRemove:    false,
	//		PullParent:     false,
	//	},
	//},
}

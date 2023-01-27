package main

import (
	"github.com/docker/docker/api/types"
	"github.com/indra-labs/indra/pkg/docker"
)

var (
	defaultPackagingContainer = "golang:1.19.4"
)

var packagingConfigurations = []docker.BuildConfiguration{
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "btcd-package",
	//	ContextFilePath: "/tmp/btcd-package.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/build/package.Dockerfile",
	//		Tags: []string{
	//			"v0.23.3",
	//		},
	//		BuildArgs: map[string]*string{
	//			"binaries_image":   strPtr("indralabs/btcd-build"),
	//			"binaries_version": strPtr("v0.23.3"),
	//			"packaging_image":  strPtr(defaultPackagingContainer),
	//			"target_name":      strPtr("btcd"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lnd-package",
		ContextFilePath: "/tmp/lnd-package.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/build/package.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
			},
			BuildArgs: map[string]*string{
				"binaries_image":   strPtr("indralabs/lnd-build"),
				"binaries_version": strPtr("v0.15.5-beta"),
				"packaging_image":  strPtr(defaultPackagingContainer),
				"target_name":      strPtr("lnd"),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "indra-package",
	//	ContextFilePath: "/tmp/indra-package.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/build/package.Dockerfile",
	//		Tags: []string{
	//			"dev",
	//		},
	//		BuildArgs: map[string]*string{
	//			"binaries_image":   strPtr("indralabs/indra-build"),
	//			"binaries_version": strPtr("dev"),
	//			"packaging_image":  strPtr(defaultPackagingContainer),
	//			"target_name":      strPtr("indra"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
}

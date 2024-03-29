package main

import (
	"git.indra-labs.org/dev/ind/pkg/docker"
	"github.com/docker/docker/api/types"
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
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "btcwallet-package",
	//	ContextFilePath: "/tmp/btcwallet-package.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/build/package.Dockerfile",
	//		Tags: []string{
	//			"v0.16.5",
	//		},
	//		BuildArgs: map[string]*string{
	//			"binaries_image":   strPtr("indralabs/btcwallet-build"),
	//			"binaries_version": strPtr("v0.16.5"),
	//			"packaging_image":  strPtr(defaultPackagingContainer),
	//			"target_name":      strPtr("btcwallet"),
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
				"v0.16.0-beta",
			},
			BuildArgs: map[string]*string{
				"binaries_image":   strPtr("indralabs/lnd-build"),
				"binaries_version": strPtr("v0.16.0-beta"),
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

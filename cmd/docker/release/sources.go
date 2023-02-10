package main

import (
	"git-indra.lan/indra-labs/indra/pkg/docker"
)

var sourceConfigurations = []docker.BuildConfiguration{
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "btcd-source",
	//	ContextFilePath: "/tmp/btcd-source.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/sources/btcd/official.Dockerfile",
	//		Tags: []string{
	//			"v0.23.3",
	//		},
	//		BuildArgs: map[string]*string{
	//			"sourcing_image": strPtr(defaultBuilderContainer),
	//			"source_url":     strPtr("https://github.com/btcsuite/btcd/releases/download"),
	//			"source_version": strPtr("v0.23.3"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "btcwallet-source",
	//	ContextFilePath: "/tmp/btcwallet-source.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/sources/btcwallet/official.Dockerfile",
	//		Tags: []string{
	//			"v0.16.5",
	//		},
	//		BuildArgs: map[string]*string{
	//			"sourcing_image": strPtr(defaultBuilderContainer),
	//			"source_url":     strPtr("https://github.com/btcsuite/btcwallet"),
	//			"source_version": strPtr("v0.16.5"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "lnd-source",
	//	ContextFilePath: "/tmp/lnd-source.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/sources/lnd/official.Dockerfile",
	//		Tags: []string{
	//			"v0.15.5-beta",
	//		},
	//		BuildArgs: map[string]*string{
	//			"sourcing_image": strPtr(defaultBuilderContainer),
	//			"source_url":     strPtr("https://github.com/lightningnetwork/lnd/releases/download"),
	//			"source_version": strPtr("v0.15.5-beta"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "indra-source",
	//	ContextFilePath: "/tmp/indra-source-local.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/sources/indra/local.Dockerfile",
	//		Tags: []string{
	//			"dev",
	//		},
	//		BuildArgs: map[string]*string{
	//			"sourcing_image": strPtr(defaultBuilderContainer),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "indra-source",
	//	ContextFilePath: "/tmp/indra-source.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/sources/indra/official.Dockerfile",
	//		Tags: []string{
	//			"v0.1.10",
	//		},
	//		BuildArgs: map[string]*string{
	//			"sourcing_image":            strPtr(defaultBuilderContainer),
	//			"source_release_url_prefix": strPtr("https://github.com/indra-labs/indra/releases/download"),
	//			"source_version":            strPtr("v0.1.10"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
}

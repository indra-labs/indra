package main

import (
	"github.com/docker/docker/api/types"
	"github.com/indra-labs/indra/pkg/docker"
)

var (
	defaultBuilderContainer = "golang:1.19.4"
)

var buildConfigurations = []docker.BuildConfiguration{
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "btcd-build",
	//	ContextFilePath: "/tmp/btcd-build.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/build/build.Dockerfile",
	//		Tags: []string{
	//			"v0.23.3",
	//		},
	//		BuildArgs: map[string]*string{
	//			"source_image":        strPtr("indralabs/btcd-source"),
	//			"source_version":      strPtr("v0.23.3"),
	//			"builder_image":       strPtr(defaultBuilderContainer),
	//			"target_name":         strPtr("btcd"),
	//			"target_build_script": strPtr("docker/build/targets/btcd.sh"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lnd-build",
		ContextFilePath: "/tmp/lnd-build.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/build/build.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
			},
			BuildArgs: map[string]*string{
				"source_image":        strPtr("indralabs/lnd-source"),
				"source_version":      strPtr("v0.15.5-beta"),
				"builder_image":       strPtr(defaultBuilderContainer),
				"target_name":         strPtr("lnd"),
				"target_build_script": strPtr("docker/build/targets/lnd.sh"),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "indra-build",
	//	ContextFilePath: "/tmp/indra-build.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/build/build.Dockerfile",
	//		Tags: []string{
	//			"dev",
	//		},
	//		BuildArgs: map[string]*string{
	//			"source_image":        strPtr("indralabs/indra-source"),
	//			"source_version":      strPtr("dev"),
	//			"builder_image":       strPtr(defaultBuilderContainer),
	//			"target_name":         strPtr("indra"),
	//			"target_build_script": strPtr("docker/build/targets/indra.sh"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     false,
	//	},
	//},
}

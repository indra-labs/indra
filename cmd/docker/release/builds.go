package main

import (
	"github.com/docker/docker/api/types"
	"github.com/indra-labs/indra/pkg/docker"
)

var buildConfigurations = []docker.BuildConfiguration{
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "btcd",
		ContextFilePath: "/tmp/btcd.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/btcd/btcd.Dockerfile",
			Tags: []string{
				"v0.23.3",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.23.3"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "btcctl",
		ContextFilePath: "/tmp/btcctl.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/btcd/btcctl.Dockerfile",
			Tags: []string{
				"v0.23.3",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.23.3"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lnd",
		ContextFilePath: "/tmp/lnd.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/lnd/lnd.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.15.5-beta"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lncli",
		ContextFilePath: "/tmp/lncli.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/lnd/lncli.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.15.5-beta"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	// docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "indra",
	//	ContextFilePath: "/tmp/indra-" + indra.SemVer + ".tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/indra/Dockerfile",
	//		Tags: []string{
	//			indra.SemVer,
	//			"latest",
	//		},
	//		BuildArgs:      map[string]*string{},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     true,
	//	},
	// },
}

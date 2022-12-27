package main

import (
	"context"
	"github.com/Indra-Labs/indra"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
	"io"
	"os"
	"time"
)

var (
	buildContextFilePath = "/tmp/indra-build.tar"
	buildOpts            = types.ImageBuildOptions{
		Tags: []string{"indra-labs/indra:" + indra.SemVer},
		Dockerfile: "docker/indra/Dockerfile",
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
	}
)

func build_image(cli *client.Client) (err error) {

	log.I.Ln("Building", buildOpts.Tags[0], "from", buildOpts.Dockerfile)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 120)
	defer cancel()

	var tar io.ReadCloser

	if tar, err = archive.TarWithOptions(".", &archive.TarOptions{}); check(err) {
		return
	}

	defer tar.Close()

	// Here build the actual docker image
	var response types.ImageBuildResponse

	if response, err = cli.ImageBuild(ctx, tar, buildOpts); check(err) {
		return
	}

	defer response.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)

	if err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stderr, termFd, isTerm, nil); check(err) {
		return
	}

	if _, err = cli.ImagesPrune(ctx, filters.NewArgs()); check(err) {
		return
	}

	return
}

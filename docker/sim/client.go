

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/Indra-Labs/indra"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
	"io"
	"io/ioutil"
	"os"
)

var (
	buildContextFilePath = "/tmp/indra-build.tar"
	buildOpts            = types.ImageBuildOptions{
		Tags: []string{
			"indralabs/indra:" + indra.SemVer,
			"indralabs/indra:latest",
		},
		Dockerfile: "docker/indra/Dockerfile",
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
	}
)

type Client struct {
	*client.Client

	ctx context.Context
}

func (cli *Client) BuildImage() (err error) {

	log.I.Ln("building", buildOpts.Tags[0], "from", buildOpts.Dockerfile)

	// Generate a tar file for docker's build context. It will contain the root of the repository's path.
	// A tar file is passed in to the docker daemon.
	var tar io.ReadCloser
	if tar, err = archive.TarWithOptions(".", &archive.TarOptions{}); check(err) {
		return
	}

	defer tar.Close()

	log.I.Ln("submitting build to docker...")

	// Submit a build to docker; with the context tar, and default options defined above.
	var response types.ImageBuildResponse
	if response, err = cli.ImageBuild(cli.ctx, tar, buildOpts); check(err) {
		return
	}

	defer response.Body.Close()

	// Generate a terminal for output
	termFd, isTerm := term.GetFdInfo(os.Stderr)

	if err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stderr, termFd, isTerm, nil); check(err) {
		return
	}

	log.I.Ln("pruning build container(s)...")

	// Prune the intermediate golang:x.xx builder container
	if _, err = cli.ImagesPrune(cli.ctx, filters.NewArgs()); check(err) {
		return
	}

	log.I.Ln("pruning successful.")
	log.I.Ln("build successful!")

	return
}

func (cli *Client) PushTags(opts types.ImagePushOptions) (err error) {

	log.I.Ln("pushing tagged images to repository...")

	var file []byte
	if file, err = ioutil.ReadFile(os.Getenv("INDRA_DOCKER_CONFIG")); check(err) {
		return
	}

	config := configfile.New("config.json")
	config.LoadFromReader(bytes.NewReader(file))

	// Generate a terminal for output
	termFd, isTerm := term.GetFdInfo(os.Stderr)

	var pushResponse io.ReadCloser

	for _, auth := range config.AuthConfigs {

		log.I.Ln("found", auth.ServerAddress)

		authConfigBytes, _ := json.Marshal(auth)
		authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)

		opts.RegistryAuth = authConfigEncoded

		for _, tag := range buildOpts.Tags {

			log.I.Ln("pushing", tag)

			if pushResponse, err = cli.ImagePush(cli.ctx, tag, opts); check(err) {
				return
			}

			if err = jsonmessage.DisplayJSONMessagesStream(pushResponse, os.Stderr, termFd, isTerm, nil); check(err) {
				return
			}

			if err = pushResponse.Close(); check(err) {
				return
			}
		}
	}

	return nil
}

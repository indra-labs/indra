package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/Indra-Labs/indra"
	log2 "github.com/Indra-Labs/indra/pkg/log"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	buildRepositoryName  = "indralabs/indra"
	buildContextFilePath = "/tmp/indra-" + indra.SemVer + ".tar"
	buildOpts            = types.ImageBuildOptions{
		Dockerfile: "docker/indra/Dockerfile",
		Tags: []string{
			buildRepositoryName + ":" + indra.SemVer,
			buildRepositoryName + ":" + "latest",
		},
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
	}
	isRelease = false
)

func SetRelease() {
	isRelease = true
}

type Builder struct {
	*client.Client
	ctx context.Context
}

func (cli *Builder) Build() (err error) {

	log.I.Ln("building", buildOpts.Tags[0], "from", buildOpts.Dockerfile)

	// If we're building a release, we should also tag stable.

	if isRelease {
		buildOpts.Tags = append(buildOpts.Tags, buildRepositoryName+":"+"stable")
	}

	// Generate a tar file for docker's release context. It will contain the root of the repository's path.
	// A tar file is passed in to the docker daemon.

	var tar io.ReadCloser

	if tar, err = archive.TarWithOptions(".", &archive.TarOptions{}); check(err) {
		return
	}

	defer tar.Close()

	// Submit a release to docker; with the context tar, and default options defined above.

	log.I.Ln("submitting release to docker...")

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

	// Prune the intermediate golang:x.xx builder container

	log.I.Ln("pruning release container(s)...")

	if _, err = cli.ImagesPrune(cli.ctx, filters.NewArgs()); check(err) {
		return
	}

	log.I.Ln("pruning successful.")
	log.I.Ln("release successful!")

	return
}

func (cli *Builder) Push(opts types.ImagePushOptions) (err error) {

	log.I.Ln("pushing tagged images to repository...")

	// Load the docker config

	var file []byte
	var config *configfile.ConfigFile

	if file, err = ioutil.ReadFile(os.Getenv("INDRA_DOCKER_CONFIG")); check(err) {
		return
	}

	config = configfile.New("config.json")

	config.LoadFromReader(bytes.NewReader(file))

	// Generate a terminal for output

	termFd, isTerm := term.GetFdInfo(os.Stderr)

	// Push the specified tags to each docker repository

	var pushResponse io.ReadCloser

	for _, auth := range config.AuthConfigs {

		log.I.Ln("found", auth.ServerAddress)

		// Generate an authentication token

		authConfigBytes, _ := json.Marshal(auth)

		opts.RegistryAuth = base64.URLEncoding.EncodeToString(authConfigBytes)

		// Pushes each tag to the docker repository.

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

	log.I.Ln("sucessfully pushed!")

	return nil
}

func NewBuilder(ctx context.Context, cli *client.Client) (builder *Builder) {

	return &Builder{
		cli,
		ctx,
	}
}

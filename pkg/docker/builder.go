package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"

	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

var (
	isRelease  = false
	isPushable = false
)

// SetRelease enables releasing of docker images.
func SetRelease() {
	isRelease = true
}

// SetPush enables pushing of dockre images.
func SetPush() {
	isPushable = true
}

// Builder is a data structure defining a docker build.
type Builder struct {
	*client.Client
	ctx            context.Context
	configs        []BuildConfiguration
	source_configs []BuildConfiguration
	pkg_configs    []BuildConfiguration
}

func (b *Builder) build(buildConfig BuildConfiguration) (err error) {

	// We need the absolute path for build tags to be valid
	buildConfig.BuildOpts.Tags = buildConfig.FixTagPrefix()

	log.I.Ln("building", buildConfig.BuildOpts.Tags[0], "from", buildConfig.BuildOpts.Dockerfile)

	// If we're building a release, we should also tag stable.

	if isRelease {
		buildConfig.BuildOpts.Tags = append(buildConfig.BuildOpts.Tags, "stable")
	}

	// Generate a tar file for docker's release context. It will contain the root of the repository's path.
	// A tar file is passed in to the docker daemon.

	var tar io.ReadCloser

	if tar, err = archive.TarWithOptions(".", &archive.TarOptions{}); check(err) {
		return
	}

	defer tar.Close()

	// Submit a release to docker; with the context tar, and default options defined above.

	log.I.Ln("running build with docker...")

	var response types.ImageBuildResponse

	if response, err = b.ImageBuild(b.ctx, tar, buildConfig.BuildOpts); check(err) {
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

	if _, err = b.ImagesPrune(b.ctx, filters.NewArgs()); check(err) {
		return
	}

	log.I.Ln("pruning successful.")
	log.I.Ln("release successful!")

	return
}

// Build runs a Builder specification.
func (b *Builder) Build() (err error) {

	for _, buildConfig := range b.source_configs {

		if err = b.build(buildConfig); check(err) {
			return
		}
	}

	for _, buildConfig := range b.configs {

		if err = b.build(buildConfig); check(err) {
			return
		}
	}

	for _, buildConfig := range b.pkg_configs {

		if err = b.build(buildConfig); check(err) {
			return
		}
	}

	return nil
}

func (b *Builder) push(buildConfig BuildConfiguration) (err error) {

	if !isPushable {
		return nil
	}

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

		buildConfig.PushOpts.RegistryAuth = base64.URLEncoding.EncodeToString(authConfigBytes)

		// Pushes each tag to the docker repository.

		for _, tag := range buildConfig.FixTagPrefix() {

			log.I.Ln("pushing", tag)

			if pushResponse, err = b.ImagePush(b.ctx, tag, buildConfig.PushOpts); check(err) {
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

// Push a build to the configured docker registry/repository.
func (b *Builder) Push() (err error) {

	for _, buildConfig := range b.configs {

		if err = b.push(buildConfig); check(err) {
			return
		}
	}

	return nil
}

// NewBuilder bundles a builder together from its parts.
func NewBuilder(ctx context.Context, cli *client.Client, sourceConfigs []BuildConfiguration,
	buildConfigs []BuildConfiguration, pkgConfigs []BuildConfiguration) (builder *Builder) {

	return &Builder{cli, ctx, buildConfigs, sourceConfigs, pkgConfigs}
}

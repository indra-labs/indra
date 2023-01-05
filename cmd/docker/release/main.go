package main

import (
	"context"
	"os"
	"time"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/docker"
	log2 "github.com/Indra-Labs/indra/pkg/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	timeout = 120 * time.Second
)

func main() {

	var err error
	var cli *client.Client

	if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); check(err) {
		os.Exit(1)
	}

	defer cli.Close()

	// Set a Timeout for 120 seconds
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	var builder = docker.NewBuilder(ctx, cli)

	if err = builder.Build(); check(err) {
		os.Exit(1)
	}

	if err = builder.Push(types.ImagePushOptions{}); check(err) {
		os.Exit(1)
	}

	if err = builder.Close(); check(err) {
		os.Exit(1)
	}

	os.Exit(0)
}

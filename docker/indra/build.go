package main

import (
	"context"
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/docker"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"os"
	"time"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func main() {

	var err error
	var cli *client.Client

	if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
		panic(err)
	}

	defer cli.Close()

	// Set a Timeout for 120 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 120)

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

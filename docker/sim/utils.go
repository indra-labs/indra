package main

import (
	"github.com/docker/docker/client"
)

func createNetworkIfNotExists(cli *client.Client) (err error) {

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second * 120)
	//defer cancel()

	//cli.NetworkCreate(ctx, "indranet")
	return
}

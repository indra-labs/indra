package main

import (
	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/docker/docker/client"
	"os"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func Setup() (err error) {

	return nil
}

func Run() (err error){

	return nil
}

func Teardown() (err error) {

	return nil
}

func main() {

	var err error
	var cli *client.Client

	if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
		panic(err)
	}

	defer cli.Close()

	build_image(cli)

	/*reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello world"},
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)*/

	os.Exit(0)
}

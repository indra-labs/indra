#!/bin/bash

docker build -t indralabs/scratch-builder ./docker/scratch/.

docker run --rm -it --volume=${PWD}/docker/scratch/tmp:/output indralabs/scratch-builder cp /tmp/root-fs.tgz /output

docker image import ${PWD}/docker/scratch/tmp/root-fs.tgz indralabs/scratch
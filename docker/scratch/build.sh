#!/bin/bash

docker build -t indralabs/scratch-builder .

docker run --rm -it --volume=${PWD}/tmp:/output indralabs/scratch-builder cp /tmp/root-fs.tgz /output

docker image import tmp/root-fs.tgz indralabs/scratch

docker push indralabs/scratch:latest
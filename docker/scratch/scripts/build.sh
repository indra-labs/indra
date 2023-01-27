#!/bin/bash

if [ -z "${TAG}" ]; then
    TAG="latest"
fi

if [ -z "${TARGET_TAG}" ]; then
    TARGET_TAG=$TAG
fi

docker build -f ./docker/scratch/root-fs.Dockerfile -t indralabs/scratch-root-fs:$TAG .

rm -rf release/scratch-$TARGET_TAG && mkdir -pv release/scratch-$TARGET_TAG

docker run --rm -it --user=$UID:$GID \
  --volume=${PWD}/release:/tmp/release indralabs/scratch-root-fs:$TAG \
    cp /tmp/root-fs.tar.gz /tmp/release/scratch-$TARGET_TAG/root-fs.tar.gz

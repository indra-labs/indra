#!/bin/bash

TARGET_REPOSITORY="indralabs/indra-multi-arch"
TARGET_TAG="dev"

IMAGES=$(docker image ls --filter="reference=$TARGET_REPOSITORY*-$TARGET_TAG" --format="{{.Repository}}:{{.Tag}}")

for i in $IMAGES; do

    echo "pushing "$i

    docker push $i

done

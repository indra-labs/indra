#!/bin/bash

MANIFEST_REPOSITORY="indralabs/indra"
TARGET_REPOSITORY="indralabs/indra-multi-arch"
TARGET_TAG="dev"

IMAGES=$(docker image ls --filter="reference=$TARGET_REPOSITORY*-$TARGET_TAG" --format="{{.Repository}}:{{.Tag}}")

echo "removing old manifest"

docker manifest rm indralabs/indra:$TARGET_TAG

CMD="docker manifest create $MANIFEST_REPOSITORY:$TARGET_TAG"

for image in $IMAGES; do
    CMD+=" --amend $image"
done

$CMD

docker manifest push $MANIFEST_REPOSITORY:$TARGET_TAG
docker manifest inspect $MANIFEST_REPOSITORY:$TARGET_TAG > release/indra-$TARGET_TAG/release/manifest-docker-$TARGET_TAG.json

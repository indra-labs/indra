#!/bin/bash

if [ -z "${TARGET_NAME}" ]; then
    echo "-- error: a release name is required to run the assembler"
    exit 1
fi

if [ -z "${TARGET_TAG}" ]; then
    echo "-- error: a release tag is required to run an extract"
    exit 1
fi

DOCKERFILES=$(find ./docker/release/targets/$TARGET_NAME -maxdepth 1 -type f -iname "*.Dockerfile" | grep -oP "(.$TARGET_NAME/\K).*")

PLATFORMS=$(find release/$TARGET_NAME-$TARGET_TAG/bin -type d | grep -oP "(^.*/bin/\K).*" | cut -f1 -d/)

if [ -z "${RELEASE}" ]; then
    PLATFORMS=(
        "linux-amd64"
        "linux-arm64"
        "linux-arm-v7"
    )
fi

echo "-- assembling images for package ${TARGET_NAME}-${TARGET_TAG}"

for DOCKERFILE in $DOCKERFILES; do

    IMAGE=$(echo $DOCKERFILE | cut -f1 -d.)
    TARGET_REPOSITORY="indralabs/$(echo $DOCKERFILE | cut -f1 -d.)-multi-arch"

    echo "-- running assembler on $IMAGE"

    docker image ls --filter="reference=$TARGET_REPOSITORY*-$TARGET_TAG" -q | xargs docker image rm -f

    for PLATFORM in $PLATFORMS; do

        echo "--- assembling $TARGET_REPOSITORY:$PLATFORM-$TARGET_TAG"

        DOCKER_PLATFORM=$(echo $PLATFORM | tr '-' '/')
        DOCKER_PLATFORM_TAG=$PLATFORM-$TARGET_TAG
        SCRATCH_PLATFORM_TAG=$PLATFORM-latest

        if [ -z "${RELEASE}" ]; then
            DOCKER_PLATFORM_TAG=$PLATFORM-dev
        fi

        docker build --quiet --platform=$DOCKER_PLATFORM \
          --build-arg platform=$PLATFORM \
          --build-arg version=$TARGET_TAG \
          --build-arg scratch_version=$SCRATCH_PLATFORM_TAG \
          -t "$TARGET_REPOSITORY:$DOCKER_PLATFORM_TAG" \
          -f ./docker/release/targets/$TARGET_NAME/$DOCKERFILE .

    done
done

#PKGS="indralabs/indra-package:dev,indralabs/btcd-package:v0.23.3,indralabs/lnd-package:v0.15.5-beta" docker/release/scripts/run.sh

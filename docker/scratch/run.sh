#!/bin/bash

PLATFORMS=${SYS:-"
  linux/386
  linux/amd64
  linux/arm/v5
  linux/arm/v6
  linux/arm/v7
  linux/arm64
  linux/loong64
  linux/mips
  linux/mips64
  linux/mips64le
  linux/mipsle
  linux/ppc64
  linux/ppc64le
  linux/riscv64
  linux/s390x
"}

if [ -z "${TAG}" ]; then
    TAG="latest"
fi

if [ -z "${TARGET_TAG}" ]; then
    TARGET_TAG=$TAG
fi

for PLATFORM in $PLATFORMS; do

  DOCKER_PLATFORM_TAG=$(echo $PLATFORM | tr '/', '-')"-$TAG"

  echo "generating image: "$PLATFORM

  touch ${PWD}/release/scratch-$TAG/arch

  echo -e "FROM scratch\nADD release/scratch-$TAG/arch /$(echo $PLATFORM | tr '/', '-').arch\nADD release/scratch-$TAG/root-fs.tar.gz /\n" | \
    docker build -t indralabs/scratch-multi-arch:$DOCKER_PLATFORM_TAG --platform=$PLATFORM -f- .

done

for PLATFORM in $PLATFORMS; do

    DOCKER_PLATFORM_TAG=$(echo $PLATFORM | tr '/', '-')"-$TAG"

    docker push indralabs/scratch-multi-arch:$DOCKER_PLATFORM_TAG

done

docker manifest rm indralabs/scratch:$TARGET_TAG

CMD="docker manifest create indralabs/scratch:$TARGET_TAG"

for PLATFORM in $PLATFORMS; do

    DOCKER_PLATFORM_TAG=$(echo $PLATFORM | tr '/', '-')"-$TAG"

    CMD+=" --amend indralabs/scratch-multi-arch:$DOCKER_PLATFORM_TAG"

done

$CMD

docker manifest push indralabs/scratch:$TARGET_TAG

#!/bin/bash

if [ -z "${RELEASE}" ]; then
    echo "overriding all tags with 'dev'. Run with RELEASE=true environment variable to override."
fi

if [ -z "${PUSH}" ]; then
    echo "push to docker repository disabled. Run with the PUSH=true environment variable to override."
fi

if [ -z "${MANIFEST}" ]; then
    echo "creation of manifest disabled. Run with the MANIFEST=true environment variable to override."
fi

if [ -z "${PKGS}" ]; then
    PKGS="indralabs/indra-package:dev"
fi

PKGS=$(echo $PKGS | tr ",", "\n")

echo "- finding packages"

for PKG in $PKGS
do

    TARGET_NAME=$(echo $PKG | grep -oP "^.*/\K(.*)(?=-package.*)")
    TARGET_TAG=$(echo $PKG | cut -f2 -d:)

    PKG_REPOSITORY="$PKG" PKG="$TARGET_NAME-$TARGET_TAG" ./docker/release/scripts/extract.sh

done

echo "- all packages found!"
echo "- finding images to assemble"

for PKG in $PKGS
do

    TARGET_NAME=$(echo $PKG | grep -oP "^.*/\K(.*)(?=-package.*)")
    TARGET_TAG=$(echo $PKG | cut -f2 -d:)

    TARGET_NAME="${TARGET_NAME}" TARGET_TAG="${TARGET_TAG}" RELEASE="${RELEASE}" ./docker/release/scripts/assemble.sh

done

if [ "${PUSH}" ]; then

  echo "- pushing packages to docker repositories"

  for PKG in $PKGS
  do

      TARGET_NAME=$(echo $PKG | grep -oP "^.*/\K(.*)(?=-package.*)")
      TARGET_TAG=$(echo $PKG | cut -f2 -d:)

      if [ -z "${RELEASE}" ]; then
          TARGET_TAG="dev"
      fi

      echo "pushing $TARGET_NAME-$TARGET_TAG"

      SERVICES=$(find ./docker/release/targets/$TARGET_NAME -maxdepth 1 -type f -iname "*.Dockerfile"  | grep -oP ".*$TARGET_NAME/\K(.*)(?=.Dockerfile)")

      for SERVICES in $SERVICES
      do
          IMAGES=$(docker image ls --filter="reference=indralabs/$SERVICES-multi-arch*$TARGET_TAG" --format="{{.Repository}}:{{.Tag}}")

          for IMAGE in $IMAGES
          do
              docker push --quiet $IMAGE
          done

      done

  done
fi

if [ "${MANIFEST}" ]; then

  echo "- pushing packages to docker repositories"

  for PKG in $PKGS
  do

      TARGET_NAME=$(echo $PKG | grep -oP "^.*/\K(.*)(?=-package.*)")
      TARGET_TAG=$(echo $PKG | cut -f2 -d:)

      RELEASE_TAG=$TARGET_TAG

      if [ -z "${RELEASE}" ]; then
          RELEASE_TAG="dev"
      fi

      SERVICES=$(find ./docker/release/targets/$TARGET_NAME -maxdepth 1 -type f -iname "*.Dockerfile"  | grep -oP ".*$TARGET_NAME/\K(.*)(?=.Dockerfile)")

      for SERVICE in $SERVICES
      do

          echo "creating manifest for indralabs/$SERVICE:$RELEASE_TAG"

          docker manifest rm indralabs/$SERVICE:$RELEASE_TAG

          CMD="docker manifest create indralabs/$SERVICE:$RELEASE_TAG"

          IMAGES=$(docker image ls --filter="reference=indralabs/$SERVICE-multi-arch*$RELEASE_TAG" --format="{{.Repository}}:{{.Tag}}")

          for image in $IMAGES; do
              CMD+=" --amend $image"
          done

          $CMD

          docker manifest push indralabs/$SERVICE:$RELEASE_TAG
          docker manifest inspect indralabs/$SERVICE:$RELEASE_TAG > release/$TARGET_NAME-$TARGET_TAG/release/manifest-docker-$SERVICE-$RELEASE_TAG.json

      done

  done
fi

#RELEASE=true PUSH=true MANIFEST=true PKGS="indralabs/indra-package:dev,indralabs/btcd-package:v0.23.3,indralabs/lnd-package:v0.15.5-beta" ./docker/release/scripts/run.sh

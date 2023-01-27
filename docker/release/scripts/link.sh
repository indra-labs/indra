
if [ -z "${RELEASE_TAG}" ]; then
    echo "a release tag is required. use RELEASE_TAG='<tag>' to continue."
    exit 1
fi

if [ -z "${PKGS}" ]; then
    PKGS="indralabs/indra-package:dev"
fi

PKGS=$(echo $PKGS | tr ",", "\n")

echo "- pushing packages to docker repositories"

for PKG in $PKGS
do

    echo "-- linking $PKG to $RELEASE_TAG"

    TARGET_NAME=$(echo $PKG | cut -f2 -d/ | cut -f1 -d:)
    TARGET_TAG=$(echo $PKG | cut -f2 -d:)

    SERVICES=$(find ./docker/release/targets/$TARGET_NAME -maxdepth 1 -type f -iname "*.Dockerfile"  | grep -oP ".*$TARGET_NAME/\K(.*)(?=.Dockerfile)")

    for SERVICE in $SERVICES
    do

        echo "-- creating manifest for indralabs/$SERVICE:$RELEASE_TAG"

        docker manifest rm indralabs/$SERVICE:$RELEASE_TAG

        CMD="docker manifest create indralabs/$SERVICE:$RELEASE_TAG"

        IMAGES=$(docker image ls --filter="reference=indralabs/$SERVICE-multi-arch*$TARGET_TAG" --format="{{.Repository}}:{{.Tag}}")

        for image in $IMAGES; do
            CMD+=" --amend $image"
        done

        $CMD

        docker manifest push indralabs/$SERVICE:$RELEASE_TAG
        docker manifest inspect indralabs/$SERVICE:$RELEASE_TAG > release/$TARGET_NAME-$TARGET_TAG/release/manifest-docker-$SERVICE-$RELEASE_TAG.json

    done

done

# RELEASE_TAG="latest" PKGS="indralabs/btcd:v0.23.3" docker/release/scripts/link.sh

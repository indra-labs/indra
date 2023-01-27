#!/bin/bash

if [ -z "${PKG}" ]; then
    echo "-- error: a pkg name is required to run an extract"
    exit 1
fi

if [ -z "${PKG_REPOSITORY}" ]; then
    echo "-- error: a pkg repository is required to run an extract"
    exit 1
fi

echo "- finding package $PKG"

if [ ! -d "release/${PKG}" ]; then

    echo "-- binaries not found in release/${PKG}"
    echo "-- checking if archive exists release/${PKG}.tar.gz"

    if [ ! -e "release/${PKG}.tar.gz" ]; then

        echo "-- archive does not exist, attempting to extract from $PKG_REPOSITORY"

        docker run --rm -it --user=$UID:$GID \
          --volume=${PWD}/release:/tmp/release $PKG_REPOSITORY \
            find /tmp/ -maxdepth 1 -name *.tar.gz -exec cp {} /tmp/release \;
    fi

      echo "-- extracting archive from release/${PKG}.tar.gz"

      mkdir release/${PKG}
      tar -xzf release/${PKG}.tar.gz --directory ./release/${PKG}
fi

echo "-- found $PKG"

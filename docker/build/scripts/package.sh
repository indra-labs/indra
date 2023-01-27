#!/bin/bash

PLATFORMS=$(find ./bin -type d | grep -oP "(^.*bin/\K).*")

# PACKAGE LOOP
for PLATFORM in $PLATFORMS; do

  echo "running packaging for $PLATFORM to release/${target_name}-$PLATFORM-${binaries_version}.tar.gz"

  tar -czvf ./release/${target_name}-$PLATFORM-${binaries_version}.tar.gz -C ./bin/$PLATFORM/. .

done

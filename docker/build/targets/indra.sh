#!/bin/bash

NAME="indra"

#PLATFORMS=(
#    "linux/386"
#    "linux/amd64"
#    "linux/arm/v5"
#    "linux/arm/v6"
#    "linux/arm/v7"
#    "linux/arm64"
##    "linux/loong64"
#    "linux/mips"
#    "linux/mips64"
#    "linux/mips64le"
#    "linux/mipsle"
#    "linux/ppc64"
#    "linux/ppc64le"
##    "linux/riscv64"
#    "linux/s390x"
#)

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm/v7"
)

# C bad
export CGO_ENABLED=0

# BUIDL LOOP
for i in "${PLATFORMS[@]}"; do

  OS=$(echo $i | cut -f1 -d/)
  ARCH=$(echo $i | cut -f2 -d/)
  ARM=$(echo $i | cut -f3 -d/)

  ARM_VARIANT=""

  if [ "$ARM" != "" ]; then
      ARM_VARIANT="-${ARM}"
  fi

  echo "running build for $OS-$ARCH$ARM_VARIANT"

  GOOS=$OS GOARCH=$ARCH GOARM=$(echo $ARM | cut -f1 -dv) go build --ldflags '-w -s' \
      -o /tmp/$NAME-${source_version}/bin/$OS-$ARCH$ARM_VARIANT/indra ./cmd/indra/.

done

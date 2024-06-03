#!/usr/bin/env bash

export DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "${DIR}" || exit 1

cd ${DIR}/../..

GOLANG_VERSION=1.22.3
UPX_VERSION=3.96-2

VERSION_SHORT=$(echo "$(cat ${DIR}/VERSION)")
PLATFORMS=linux/arm/v7,linux/arm64/v8,linux/amd64

function build_image() {
  FILE=$1
  shift
  TAGS=""
  for var in "$@"
  do
      TAGS="${TAGS} -t slinkgo/void:$var"
  done
  echo "FILE: $FILE"
  echo "TAGS: $TAGS"
  docker buildx create --use
  docker buildx build                                 \
     -f ${DIR}/Dockerfile-$FILE                       \
     --push                                           \
     --platform ${PLATFORMS}                          \
     --build-arg "GOLANG_VERSION=${GOLANG_VERSION}"   \
     --build-arg "UPX_VERSION=${UPX_VERSION}"         \
     ${TAGS} .
     docker buildx rm
}


build_image "gin" "$VERSION_SHORT" "latest"

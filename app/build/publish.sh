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
  IMAGE=$1
  shift
  TYPE=$1
  shift
  TAGS=""
  for var in "$@"
  do
      TAGS="${TAGS} -t slinkgo/${IMAGE}:$var"
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
     --build-arg "TYPE=${TYPE}"                \
     ${TAGS} .
     docker buildx rm
}

if [ "$1" == "gw" ]; then
  build_image "gin" "void" "" "$VERSION_SHORT" "latest"
elif [ "$1" == "back" ]; then
  build_image "back" "test-backend" "gin" "0.0.1-gin" "latest"
else
  echo "Usage: ./publish.sh { gw | back }"
fi


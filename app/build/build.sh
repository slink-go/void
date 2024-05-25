#!/usr/bin/env bash

GOLANG_VERSION=1.22.3
UPX_VERSION=3.96-2

export DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "${DIR}" || exit 1

VERSION_SHORT=$(echo "$(cat ${DIR}/VERSION)")
VERSION_LONG=$(echo "v$(cat ${DIR}/VERSION) ($(git describe --abbrev=8 --always --long))")
#cat ${DIR}/build/logo.txt | sed -e "s/0.0.0/${VERSION_LONG}/g" > ${DIR}/src/logo.txt

cd ${DIR}/../..

function build_backend() {
    docker build -f ${DIR}/Dockerfile-back                      \
               --tag mvkvl/api-backend:${VERSION_SHORT}-$1      \
               --build-arg "TYPE=$1"                            \
               --build-arg "GOLANG_VERSION=${GOLANG_VERSION}"   \
               --build-arg "UPX_VERSION=${UPX_VERSION}" .
}
function build_gw_fiber() {
  docker build -f ${DIR}/Dockerfile-fiber                       \
               --tag mvkvl/api-gateway-fiber:${VERSION_SHORT}   \
               --build-arg "GOLANG_VERSION=${GOLANG_VERSION}"   \
               --build-arg "UPX_VERSION=${UPX_VERSION}" .
}

if [[ "$1" == "-h" || "$1" == "help" ]]; then
  echo "Usage: ./build.sh [ back | gin | fiber ]"
  exit 1
elif [ "$1" == "fiber" ]; then
  build_gw_fiber
elif [ "$1" == "back" ]; then
  build_backend "gin"
#  build_backend "fiber"
else
  build_gw_fiber
  build_backend "gin"
#  build_backend "fiber"
fi
docker system prune -f
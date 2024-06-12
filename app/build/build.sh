#!/usr/bin/env bash

GOLANG_VERSION=1.22.3
UPX_VERSION=3.96-2

export DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "${DIR}" || exit 1

DISTR="${DIR}/../distr"

VERSION_SHORT=$(echo "$(cat ${DIR}/VERSION)")
VERSION_LONG=$(echo "v$(cat ${DIR}/VERSION) ($(git describe --abbrev=8 --always --long))")
#cat ${DIR}/build/logo.txt | sed -e "s/0.0.0/${VERSION_LONG}/g" > ${DIR}/src/logo.txt

cd "${DIR}/../.."

function build_backend() {
    docker build -f ${DIR}/Dockerfile-back                      \
               --tag mvkvl/api-backend:${VERSION_SHORT}-$1      \
               --build-arg "TYPE=$1"                            \
               --build-arg "GOLANG_VERSION=${GOLANG_VERSION}"   \
               --build-arg "UPX_VERSION=${UPX_VERSION}" .
}
function build_gw() {
  docker build -f ${DIR}/Dockerfile-$1                          \
               --tag mvkvl/api-gateway-$1:${VERSION_SHORT}      \
               --build-arg "GOLANG_VERSION=${GOLANG_VERSION}"   \
               --build-arg "UPX_VERSION=${UPX_VERSION}" .
}
function build_bin() {
  OS=$1
  ARCH=$2
  OUTDIR=$3
  OUT=$4
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags="-s -w" -o ${DISTR}/$OUTDIR/$OUT .
  cwd=$(pwd)
  cd ${DISTR}/$OUTDIR
  tar cvfz "${DISTR}/${OUT%.*}_${VERSION_SHORT}_${OS}_${ARCH}.tgz" ${OUT}
  cd $cwd
}
function build_dir() {
  SRCDIR="$1"
  NAME="$2"
  cd "${SRCDIR}"
  go get -u
  go mod tidy
  mkdir -p ${DISTR}/$NAME/macos/{amd64,arm64}
  mkdir -p ${DISTR}/$NAME/linux/amd64
  mkdir -p ${DISTR}/$NAME/windows/amd64
  build_bin darwin  amd64 "$NAME/macos/amd64" "${NAME}"
  build_bin darwin  arm64 "$NAME/macos/arm64" "${NAME}"
  build_bin linux   amd64 "$NAME/linux/amd64" "${NAME}"
  build_bin linux   arm64 "$NAME/linux/arm64" "${NAME}"
  build_bin windows amd64 "$NAME/windows/amd64" "${NAME}.exe"
  rm -rf ${DISTR}/$NAME 2> /dev/null
}
function docker_login() {
  source "${DIR}/.env" || true
  if [ -z "$DOCKER_LOGIN" ]; then
    echo "DOCKER_LOGIN variable not set"
    exit 1
  fi
  if [ -z "$DOCKER_PASSWORD" ]; then
    echo "DOCKER_PASSWORD variable not set"
    exit 1
  fi
  echo "$DOCKER_PASSWORD" | docker login --password-stdin -u "$DOCKER_LOGIN"
}

if [[ "$1" == "-h" || "$1" == "help" ]]; then
  echo "Usage: ./build.sh [ back | gin | bin ]"
  exit 1
elif [ "$1" == "gin" ]; then
  docker_login
  build_gw "gin"
elif [ "$1" == "back" ]; then
  docker_login
  build_backend "gin"
elif [ "$1" == "bin" ]; then
  rm -rf   ${DIR}/../distr 2> /dev/null
  build_dir "${DIR}/../../api-gateway/cmd/gin" "void"
  build_dir "${DIR}/../../backend/gin" "backend"
  exit 0
else
  docker_login
  build_gw "gin"
  build_backend "gin"
fi
docker system prune -f

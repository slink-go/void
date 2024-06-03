#!/usr/bin/env bash

export DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "${DIR}"/../../api-gateway || exit 1

deadcode ./...
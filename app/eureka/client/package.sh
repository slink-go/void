#    --platform $(platform_param)   \

DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

if [ -n "$1" ]; then
  TARGET="--target=$1"
fi

cd ${DIR}
docker build -t "mvkvl/eureka-client" ${TARGET} .

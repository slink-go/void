#    --platform $(platform_param)   \

if [ -n "$1" ]; then
  TARGET="--target=$1"
fi

docker build -t "mvkvl/eureka" ${TARGET} .


COMPOSE="docker compose"

if [ "$1" == "--pull" ]; then
  ${COMPOSE} pull
fi

if [ "$1" == "-f" ]; then
  ${COMPOSE} up
else
  ${COMPOSE} up -d
fi

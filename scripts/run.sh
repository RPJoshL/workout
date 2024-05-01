# Color definitions
GREEN='\033[0;32m'
NC='\033[0m'

# App settings
export LOGGER_LEVEL="DEBUG"
export SERVER_FQDN="http://localhost:4020"
export DEV_MODE=true

# Set secret variables
export $(cat ./scripts/secrets | xargs)

# Set module to run
module="./cmd/workout"
if [ "$1" == "ddl" ]; then
	go run ./cmd/ddl
	exit 0
fi
if [ "$1" == "geonames" ]; then
	go run ./cmd/geonames
	exit 0
fi

# Run app
nodemon --delay 0.1s -e go,html,yaml,templ,css,scss,js -i '*_templ.go' -i 'pages.css' -i 'pages.scss' --signal SIGTERM --quiet --exec \
'echo -e "\n'"$GREEN"'[Restarting]'"$NC"'" && templ generate > /dev/null 2>&1 || true && make css > /dev/null 2>&1 || true && go run '"$module" -- "$@" || exit 1""

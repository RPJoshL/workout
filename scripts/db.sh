#!/bin/bash

# =========================================================
# This script creates a test database for unit testing
# =========================================================

CONTAINER_NAME="workout-test-mariadb"
CONTAINER_ID=""

# Tries to fetch the ID of a running mariadb container
getInstance() {
	CONTAINER_ID="$(podman ps | grep ""$CONTAINER_NAME"" | cut -d ' ' -f1)"
}

# Executes a single statement on the Mariadb docker container
executeCommand() {
	podman exec -i "$CONTAINER_ID" mariadb -u root -p"test-driven" -e "$1"
}

# Executes a SQL script inside ./db/migrations/ for the specified db
executeScript() {
	# Piping into podman does NOT WORK ('|' or '<')
	podman cp "./db/migrations/$1" "$CONTAINER_ID:/"
	podman exec -i "$CONTAINER_ID" mariadb -u root -p"test-driven" "$DB_DB" -e "source /"$1""
}

# Get any running container
getInstance

# Check non default flags
if [ "$1" == "delete" ] || [ "$1" == "stop" ] || [ "$1" == "rm" ]; then
	if [ "$CONTAINER_ID" = "" ]; then
		echo "No container running"
		exit 1
	fi

	# Stop it
	podman stop "$CONTAINER_ID"
	podman rm "$CONTAINER_ID"
	exit 0
fi

# Nothing to do if container exists
if [ "$CONTAINER_ID" != "" ]; then
	echo "Container is already running"
	exit 0
fi

# Create a new container
podman run --detach --name "$CONTAINER_NAME" \
	--env MARIADB_ROOT_PASSWORD=test-driven \
	-p 3306:3306 docker.io/mariadb:11.3
getInstance
sleep 5

# Source test secrets out
export $(cat ./scripts/secrets | xargs)
export $(cat ./scripts/secrets_test | xargs)

# Create default schema and user to operate on
executeCommand '
CREATE DATABASE '$DB_DB'
	CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
'
executeCommand '
CREATE OR REPLACE USER
	'$DB_USER'@'"'%'"' IDENTIFIED BY '"'$DB_PASSWORD'"';
'
executeCommand "GRANT ALL PRIVILEGES ON $DB_DB.* TO '$DB_USER'@'%';"

# Create schemas
for script in ./db/migrations/*.sql; do
	executeScript "$(basename $script)"
done

# Create geodata
export $(cat ./scripts/secrets_test | xargs)
# Not relevant for tests
export DISABLE_COUNTRIES=true
go run ./cmd/geonames

## Test command for unning container
## podman exec -it workout-test-mariadb mariadb -u root -p"test-driven"
#!/bin/bash

# =========================================================
# This script creates a test database for unit testing
# =========================================================

CONTAINER_NAME="workout-test-mariadb"
CONTAINER_ID=""

# Tries to fetch the ID of a running mariadb container
getInstance() {
	CONTAINER_ID="$(podman ps -a | grep ""$CONTAINER_NAME"" | cut -d ' ' -f1)"
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

# Source test secrets out
export $(cat ./scripts/secrets | xargs)
export $(cat ./scripts/secrets_test | xargs)

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
elif [ "$1" = "exec" ]; then
	if [ "$CONTAINER_ID" = "" ]; then
		echo "No container running"
		exit 1
	fi

	podman exec -it "$CONTAINER_ID" mariadb -u root -p"test-driven" --database "$DB_DB"
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
	-p 3306:3306 docker.io/mariadb:11.3 \
	--innodb-autoinc-lock-mode=2
getInstance

# Wait until database is ready
while [ 1 = 1 ]; do
	sleep 1
	podman exec -it "$CONTAINER_ID" /usr/local/bin/healthcheck.sh --su-mysql --connect --innodb_initialized > /dev/null
	if [ $? -eq 0 ]; then
		echo "Database is up and running"
		break
	fi
done

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
for script in $(ls ./db/migrations/*.sql | sort -V | grep -v "~"); do
	# echo "Executing script "$(basename $script)""
	executeScript "$(basename $script)"
done

# Create geodata
if [ ! -f ./dependencies/cities1000.txt ]; then
	echo "Warning: cities1000.txt were not downloaded"
else
	export DISABLE_COUNTRIES=true
	go run ./cmd/geonames
fi

# Create data for tests
executeCommand "SET time_zone = '+00:00'; USE workout; CALL PopulateYearDay('2020-01-01', '2025-12-31');"

## Test command for unning container
## podman exec -it workout-test-mariadb mariadb -u root -p"test-driven"
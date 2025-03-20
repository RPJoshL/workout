# Import deploy configuration
dpl ?= deploy.env
# Check if file exists
ifeq ($(wildcard $(dpl)),)
$(info File $(dpl) does not exist. No variables are applied.)
else
include $(dpl)
export $(shell sed 's/=.*//' $(dpl))
endif

# Get the current version
VERSION=$(shell cat ./VERSION)
WORKDIR=$(shell pwd)
UID=$(shell echo $uid)


.PHONY: help

# Output help for every task
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

setup: install-dev install-js install-css install-dependencies ## Installs all dependencies needed to run templ

install-dev: ## Installs development tools needed to run this application
	go install github.com/a-h/templ/cmd/templ@v0.3.833
	sudo cp ${HOME}/go/bin/templ /usr/bin
	sudo npm install -g sass-embedded
	sudo npm install -g nodemon
	sudo npm install -g minify
	sudo npm install -g typescript@5.4.5

install-js: ## Installs required javascript dependencies
	npm ci --include=dev

	rm -rf ./static/js/3dparty/*.js
	rm -rf ./node_modules/@types/leaflet-fullscreen ./node_modules/@types/leaflet-geometryutil
	mkdir -p ./node_modules/@types/leaflet-fullscreen ./node_modules/@types/leaflet-geometryutil

	# HTMX
	wget https://unpkg.com/htmx.org@1.9.11 -O ->> ./static/js/3dparty/main.js
	wget https://unpkg.com/htmx.org@1.9.11/dist/ext/response-targets.js -O - | minify --js | tee >> ./static/js/3dparty/main.js
	
	# Toastify is modified locally
	# wget https://cdn.jsdelivr.net/npm/toastify-js@1.12.0 -O ->> ./static/js/3dparty/main.js
	minify ./static/js/toastify.js >> ./static/js/3dparty/main.js

	# EasyMDE (Markdown editor and viewer)
	wget https://unpkg.com/easymde@2.18.0/dist/easymde.min.js -O ->> ./static/js/3dparty/main.js

	# Leaflet
	wget https://unpkg.com/leaflet@1.9.4/dist/leaflet.js -O ->> ./static/js/3dparty/main.js
	wget https://cdnjs.cloudflare.com/ajax/libs/leaflet-contextmenu/1.4.0/leaflet.contextmenu.min.js -O - | sed '2 i\/*' >> ./static/js/3dparty/main.js 
	wget https://raw.githubusercontent.com/runette/Leaflet.fullscreen/gh-pages/dist/Leaflet.fullscreen.min.js -O ->> ./static/js/3dparty/main.js
	wget https://raw.githubusercontent.com/runette/Leaflet.fullscreen/gh-pages/index.d.ts -O ->> ./node_modules/@types/leaflet-fullscreen/index.d.ts
	# wget https://raw.githubusercontent.com/trafficonese/Leaflet.glify/hoverOff_Shapes/dist/glify-browser.js -O ->> ./static/js/3dparty/main.js
	wget https://unpkg.com/leaflet-geometryutil@0.10.3/src/leaflet.geometryutil.d.ts -O ->> ./node_modules/@types/leaflet-geometryutil/index.d.ts
	wget https://unpkg.com/leaflet-geometryutil@0.10.3/src/leaflet.geometryutil.js -O - | tee >> ./static/js/3dparty/main.js

	# Apache echarts
	wget https://raw.githubusercontent.com/RPJoshL/echarts/master/dist/echarts.min.js -O ->> ./static/js/3dparty/main.js
	# wget https://cdn.jsdelivr.net/npm/echarts@5.5.0/dist/echarts.min.js -O ->> ./static/js/3dparty/main.js

	# Copy country flags
	cp ./node_modules/country-flag-icons/3x2/* ./static/img/svg/country-flags/

	# Flatpickr (Datepicker)
	wget https://cdn.jsdelivr.net/npm/flatpickr@4.6.13 -O ->> ./static/js/3dparty/main.js

	@HASH=$$(cat ./static/js/3dparty/main.js | sha256sum | cut -c1-16); \
		mv ./static/js/3dparty/main.js "./static/js/3dparty/main-$$HASH.js"
	
install-css: ## Installs required css dependencies
	rm -rf ./static/css/third.css

	# EasyMDE (Markdown editor and viewer)
	wget https://unpkg.com/easymde@2.18.0/dist/easymde.min.css -O ->> ./static/css/third.css

	# Toastify styles 
	wget https://cdn.jsdelivr.net/npm/toastify-js/src/toastify.min.css -O ->> ./static/css/third.css

	# Leaflet styles
	wget https://unpkg.com/leaflet@1.9.4/dist/leaflet.css -O ->> ./static/css/third.css
	wget https://cdnjs.cloudflare.com/ajax/libs/leaflet-contextmenu/1.4.0/leaflet.contextmenu.min.css -O ->> ./static/css/third.css
	wget https://raw.githubusercontent.com/runette/Leaflet.fullscreen/gh-pages/dist/Leaflet.fullscreen.min.css -O ->> ./static/css/third.css

	# Flatpickr (Datepicker)
	wget https://cdn.jsdelivr.net/npm/flatpickr@4.6.13/dist/themes/dark.min.css -O ->> ./static/css/third.css

install-dependencies: ## Install required third party dependencies
	rm -rf ./dependencies/
	mkdir ./dependencies

	# Cities
	wget https://download.geonames.org/export/dump/cities1000.zip -O ./dependencies/cities.zip
	unzip dependencies/cities.zip -d dependencies/
	rm dependencies/cities.zip

	# Full geonames of all countries
	wget https://download.geonames.org/export/dump/allCountries.zip -O ./dependencies/countries.zip
	unzip dependencies/countries.zip -d dependencies/
	rm ./dependencies/countries.zip

run: ## Runs the application in dev mode
	@./scripts/run.sh

run-modules: ## Runs and watch typescript modules for changes
	@./scripts/run.sh modules

run-uploader: ## Runs the file system watcher to upload new workouts automatically
	@./scripts/run.sh uploader

run-container:  ## Run the application within previously build container
	@ make stop-container > /dev/null 2>&1 || true
	@ podman run -it --name rpout --userns=keep-id --cap-drop ALL -p 40001:40001 \
		--env-file './scripts/secrets'  -e SERVER_ADDRESS=localhost:40001 \
		git.rpjosh.de/rpout:v$(VERSION)-dev

run-db: ## Runs a test database to perform some tests
	@./scripts/db.sh

run-android:  ## Run the container with the build android APKs
	@ make stop-android > /dev/null 2>&1 || true
	@ podman run -it --name rpout-android --userns=keep-id -p 8090:8090 -e PORT=8090 \
		git.rpjosh.de/rpout-android:v$(VERSION)-dev 
stop-android:  ## Stop and remove a previous started container with the android APKs
	@ podman stop rpout-android; podman rm rpout-android

stop-db: ## Stop the test databse
	@./scripts/db.sh stop

stop-container: ## Stop and removes a previously started container
	@ podman stop rpout; podman rm rpout

exec-db: ## Excecutes an interactive SQL shell for a previously started DB
	@./scripts/db.sh exec

css: ## Compiles all CSS files
	@LOGGER_LEVEL=DEBUG \
	 REMOVE_SCSS_FILE=FALSE \
		go run ./cmd/css

geonames: ## Imports previously downloaded geonames dumps into the db
	@./scripts/run.sh geonames

ddl: ## Generates DDL definitions for database tables
	@./scripts/run.sh ddl

modules: ## Generates JS modules
	@go run ./cmd/modules

build: ## Build a container image (with cache)
	buildah bud --layers --build-arg VERSION="$(VERSION)" \
		--secret id=giteaSshKey,src=$(GIT_SSH_KEY) \
		--tag=git.rpjosh.de/rpout:v$(VERSION)-dev \
		-f docker/server/Dockerfile .

build-customizer: ## Build the android APKs
	buildah bud --layers --build-arg VERSION="$(VERSION)" \
		--secret id=androidKeystore,src=$(ANDROID_KEYSTORE_PATH) \
		--secret id=androidKeystorePassword,src=$(ANDROID_KEYSTORE_PASSWORD) \
		--tag=git.rpjosh.de/rpout-android:v$(VERSION)-dev \
		--memory 2200m \
		-f docker/android/Dockerfile .

clear-images: ## Remove all previously build images and all intermediate images created by this makefile
	podman rmi $$(podman images -a | grep -e '<none>' -e '\/rpout-.*' | awk '{ print $3 }') -f

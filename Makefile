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

instell-dev: ## Installs development tools needed to run this application
	go install github.com/a-h/templ/cmd/templ@v0.2.680
	sudo cp ${HOME}/go/bin/templ /usr/bin
	sudo npm install -g node-sass
	sudo npm install -g nodemon
	sudo npm install -g minify
	sudo npm install -g typescript

install-js: ## Installs required javascript dependencies
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
	wget https://raw.githubusercontent.com/trafficonese/Leaflet.glify/hoverOff_Shapes/dist/glify-browser.js -O ->> ./static/js/3dparty/main.js
	wget https://unpkg.com/leaflet-geometryutil@0.10.3/src/leaflet.geometryutil.d.ts -O ->> ./node_modules/@types/leaflet-geometryutil/index.d.ts
	wget https://unpkg.com/leaflet-geometryutil@0.10.3/src/leaflet.geometryutil.js -O - | tee >> ./static/js/3dparty/main.js

	# Apache echarts
	wget https://raw.githubusercontent.com/RPJoshL/echarts/master/dist/echarts.min.js -O ->> ./static/js/3dparty/main.js
	# wget https://cdn.jsdelivr.net/npm/echarts@5.5.0/dist/echarts.min.js -O ->> ./static/js/3dparty/main.js

	# Copy country flags
	cp ./node_modules/country-flag-icons/3x2/* ./static/img/svg/country-flags/

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


install-dependencies: ## Install required third party dependencies
	rm -rf ./dependencies/
	mkdir ./dependencies

	wget https://download.geonames.org/export/dump/cities1000.zip -O ./dependencies/cities.zip
	unzip dependencies/cities.zip -d dependencies/
	rm dependencies/cities.zip


run: ## Runs the application in dev mode
	@./scripts/run.sh

run-modules: ## Runs and watch typescript modules for changes
	@./scripts/run.sh modules

run-container:  ## Run the application within previously build container
	@ make stop-container > /dev/null 2>&1 || true
	@ podman run -it --name rpout --userns=keep-id --cap-drop ALL -p 40001:40001 \
		--env-file './scripts/secrets'  -e SERVER_ADDRESS=localhost:40001 \
		git.rpjosh.de/rpout:v$(VERSION)-dev

run-db: ## Runs a test database to perform some tests
	@./scripts/db.sh

stop-db: ## Stop the test databse
	@./scripts/db.sh stop

stop-container: ## Stop and removes a previously started container
	@ podman stop rpout; podman rm rpout

css: ## Compiles all CSS files
	@LOGGER_LEVEL=DEBUG \
	 REMOVE_SCSS_FILE=FALSE \
		go run ./cmd/css

geonames: ## Imports previously downloaded geonames into the db
	@./scripts/run.sh geonames

ddl: ## Generate ddl structs
	@./scripts/run.sh ddl

modules: ## Generates JS modules
	@go run ./cmd/modules

build: ## Build a container image (with cache)
	buildah bud --layers --build-arg VERSION="$(VERSION)" \
		--secret id=giteaSshKey,src=$(GIT_SSH_KEY) \
		--tag=git.rpjosh.de/rpout:v$(VERSION)-dev \
		-f docker/Dockerfile .

clear-images: ## Remove all previously build images and all intermediate images created by this makefile
	podman rmi $$(podman images -a | grep -e '<none>' -e '\/rpout-.*' | awk '{ print $3 }') -f

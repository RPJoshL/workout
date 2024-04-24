.PHONY: help

# Output help for every task
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

setup: ## Installs all dependencies needed to run templ
	go install github.com/a-h/templ/cmd/templ@v0.2.648
	sudo cp ${HOME}/go/bin/templ /usr/bin
	sudo npm install -g node-sass
	sudo npm install -g nodemon
	sudo npm install -g minify

install-js: ## Installs required javascript dependencies
	rm -rf ./static/js/3dparty/*.js

	# HTMX
	wget https://unpkg.com/htmx.org@1.9.11 -O ->> ./static/js/3dparty/main.js
	wget https://unpkg.com/htmx.org@1.9.11/dist/ext/response-targets.js -O ->> ./static/js/3dparty/main.js
	
	# Toastify is modified locally
	# wget https://cdn.jsdelivr.net/npm/toastify-js@1.12.0 -O ->> ./static/js/3dparty/main.js
	cat ./static/js/toastify.js >> ./static/js/3dparty/main.js

	# EasyMDE (Markdown editor and viewer)
	wget https://unpkg.com/easymde@2.18.0/dist/easymde.min.js -O ->> ./static/js/3dparty/main.js

	@HASH=$$(cat ./static/js/3dparty/main.js | sha256sum | cut -c1-16); \
		mv ./static/js/3dparty/main.js "./static/js/3dparty/main-$$HASH.js"
	

run: ## Runs the application in dev mode
	@./scripts/run.sh

run-db: ## Runs a test database to perform some tests
	@./scripts/db.sh

stop-db: ## Stop the test databse
	@./scripts/db.sh stop

install-css: ## Installs required css dependencies
	rm -rf ./static/css/third.css

	# EasyMDE (Markdown editor and viewer)
	wget https://unpkg.com/easymde@2.18.0/dist/easymde.min.css -O ->> ./static/css/third.css

	# Toastify styles 
	wget https://cdn.jsdelivr.net/npm/toastify-js/src/toastify.min.css -O ->> ./static/css/third.css


css: ## Compiles all CSS files
	@LOGGER_LEVEL=DEBUG \
	 REMOVE_SCSS_FILE=FALSE \
		go run ./cmd/css

ddl: ## Generate ddl structs
	@./scripts/run.sh ddl
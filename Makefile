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

run: ## Runs the application in dev mode
	@./scripts/run.sh

css: ## Compiles all CSS files
	@LOGGER_LEVEL=DEBUG \
	 REMOVE_SCSS_FILE=FALSE \
		go run ./cmd/css
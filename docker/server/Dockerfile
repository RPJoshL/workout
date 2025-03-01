# Buid site
FROM docker.io/golang:1.24-alpine3.21 AS go-builder

# Install required dependencies
ENV GOBIN=/usr/local/bin/
RUN go install github.com/a-h/templ/cmd/templ@v0.3.833 \
	&& apk update && apk add --no-cache nodejs npm make tzdata \
	&& npm install -g sass-embedded \
	&& npm install -g minify \
	&& npm install -g typescript@5.4.5

# Add git with credentials to support private repositories in "go get"
ENV GIT_SSH_COMMAND="ssh -i /run/secrets/giteaSshKey -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
RUN apk update && apk add --no-cache git openssh \
    && git config --global url."ssh://git@git.rpjosh.de:123".insteadOf "https://git.rpjosh.de" \
    && go env -w GOPRIVATE="git.rpjosh.de"

# Only copy and install the dependencies to support cache
WORKDIR /build
COPY ./go.sum ./go.mod ./Makefile ./.minify.json ./tsconfig.json ./package-lock.json ./
COPY ./static/ ./static/
RUN --mount=type=secret,id=giteaSshKey,required=true,mode=400 go mod download \
	&& make install-css

# Javascript dependencies
COPY ./.minify.json ./tsconfig.json ./package-lock.json ./package.json ./
RUN npm ci \
	&& make install-js

# Copy all files and build binary
COPY ./ .
RUN --mount=type=secret,id=giteaSshKey,required=true,mode=400 \
	   templ generate \
	&& go run ./cmd/css \
	&& mv ./static/css/pages.css ./static/css/pages-"$(cat ./static/css/pages.css | sha256sum | cut -c1-16)".css \
	&& go run ./cmd/modules \
	&& GOOS=linux GOARCH=amd64 go build -o "workout" ./cmd/workout


# Main container image
FROM alpine:3.21

# Copy build binary
WORKDIR /app
COPY --from=go-builder /build/workout /app/rpout
COPY ./docker/server/entrypoint.sh /entrypoint.sh

# Use some random user
USER 12345

ENTRYPOINT [ "/entrypoint.sh" ]
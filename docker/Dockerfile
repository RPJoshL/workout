# Buid site
FROM docker.io/golang:1.22-alpine3.19 AS go-builder

# Install required dependencies
ENV GOBIN=/usr/local/bin/
RUN go install github.com/a-h/templ/cmd/templ@v0.2.663 \
	&& apk update && apk add --no-cache nodejs npm make \
	&& npm install -g node-sass \
	&& npm install -g minify

# Add git with credentials to support private repositories in "go get"
ENV GIT_SSH_COMMAND="ssh -i /run/secrets/giteaSshKey -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
RUN apk update && apk add --no-cache git openssh \
    && git config --global url."ssh://git@git.rpjosh.de:123".insteadOf "https://git.rpjosh.de" \
    && go env -w GOPRIVATE="git.rpjosh.de"

# Only copy and install the dependencies to support cache
WORKDIR /build
COPY ./go.sum ./go.mod ./Makefile ./
COPY ./static/ ./static/
RUN --mount=type=secret,id=giteaSshKey,required=true,mode=400 go mod download \
	&& make install-js \
	&& make install-css

# Copy all files and build binary
COPY ./ .

RUN --mount=type=secret,id=giteaSshKey,required=true,mode=400 \
	   templ generate \
	&& go run ./cmd/css \
	&& mv ./static/css/pages.css ./static/css/pages-"$(cat ./static/css/pages.css | sha256sum | cut -c1-16)".css \
	&& GOOS=linux GOARCH=amd64 go build -o "workout" ./cmd/workout


# Main container image
FROM alpine:3.19

# Copy build binary
WORKDIR /app
COPY --from=go-builder /build/workout /app/rpout
COPY ./docker/entrypoint.sh /entrypoint.sh

# Use some random user
USER 12345

ENTRYPOINT [ "/entrypoint.sh" ]
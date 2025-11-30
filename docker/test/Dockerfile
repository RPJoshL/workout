FROM docker.io/golang:1.24-alpine3.21 

ARG SECRETS_PATH=./scripts/_secrets_ci

# Install required dependencies
ENV GOBIN=/usr/local/bin/
RUN go install github.com/a-h/templ/cmd/templ@v0.3.906

# Add git with credentials to support private repositories in "go get"
ENV GIT_SSH_COMMAND="ssh -i /run/secrets/giteaSshKey -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
RUN apk update && apk add --no-cache git openssh make \
    && git config --global url."ssh://git@git.rpjosh.de:123".insteadOf "https://git.rpjosh.de" \
    && go env -w GOPRIVATE="git.rpjosh.de"

# Only copy and install the dependencies to support cache
WORKDIR /build
COPY ./go.sum ./go.mod ./Makefile ./
RUN --mount=type=secret,id=giteaSshKey,required=true,mode=400 go mod download

# Copy all files and run tests
COPY ./ .
RUN --mount=type=secret,id=giteaSshKey,required=true,mode=400 \
	   set -a && . ${SECRETS_PATH} && set +a && \ 
	   make templ && set -o pipefail && \
	   LOGGER_LEVEL="error" go test ./... | { grep -v 'no test files'; true; }
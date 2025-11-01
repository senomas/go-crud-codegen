ARG DOCKER_REGISTRY=docker.io

FROM ${DOCKER_REGISTRY}/python:3.12-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends git make ca-certificates curl \
  && rm -rf /var/lib/apt/lists/*

# docker CLI (talks to host daemon via mounted /var/run/docker.sock)
ARG DOCKER_CLI_VERSION=27.3.1
RUN curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_CLI_VERSION}.tgz \
  | tar -xz -C /usr/local/bin --strip-components=1 docker/docker

# Install essential tools
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
  git make ca-certificates curl tar bash gcc g++ \
  && rm -rf /var/lib/apt/lists/*

# Install Go (1.25.1)
# --------------------------------------------------------------------
ARG GO_VERSION=1.25.1
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
  | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

# Optional: set Go environment defaults
ENV GOPATH=/go
ENV GOCACHE=/go/cache
ENV PATH="$GOPATH/bin:${PATH}"

RUN pip install docker cryptography

# Verify tools (for debugging)
RUN go version && python3 --version && git --version && docker --version && make --version

WORKDIR /work


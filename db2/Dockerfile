ARG DOCKER_REGISTRY=docker.io
ARG BUILD_UTIL_VER=ver
ARG PYTHON_VERSION=3.12
FROM ${DOCKER_REGISTRY}/codegen-build-util:${BUILD_UTIL_VER} AS builder

ENV GOPATH=/work/.go
ENV GOCACHE=/work/.go/cache

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go .

RUN go build -o codegen .

FROM ${DOCKER_REGISTRY}/python:${PYTHON_VERSION}-slim

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
ENV GOPATH=/work/.go
ENV GOCACHE=/work/.go/cache
ENV PATH="$GOPATH/bin:${PATH}"

# Verify tools (for debugging)
RUN go version && python3 --version && git --version && docker --version && make --version

WORKDIR /work/codegen

ARG GOIMPORTS_VER=v1

RUN go install golang.org/x/tools/cmd/goimports@${GOIMPORTS_VER}

ARG SWAGGO_VER=v1
RUN go install github.com/swaggo/swag/cmd/swag@${SWAGGO_VER}

RUN goimports -v
RUN swag -v

COPY --from=builder /app/codegen /bin/codegen

ENV IBM_DB_HOME=/opt/ibm/clidriver
ENV LD_LIBRARY_PATH=${IBM_DB_HOME}/lib
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-I${IBM_DB_HOME}/include"
# add rpath so test binaries find libdb2.so at runtime
ENV CGO_LDFLAGS="-L${IBM_DB_HOME}/lib -Wl,-rpath,${IBM_DB_HOME}/lib"

WORKDIR /work/app

WORKDIR /work/codegen
COPY base ./base
COPY db2 ./db2
COPY sqlite ./sqlite
COPY postgresql ./postgresql

COPY postgresql ./postgresql

COPY entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

WORKDIR /work/app


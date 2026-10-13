# syntax=docker/dockerfile:1

### 1. Build Webapp
FROM node:18-bookworm AS nodebuild

WORKDIR /webapp

COPY webapp/package.json webapp/package-lock.json* ./

RUN CPPFLAGS="-DPNG_ARM_NEON_OPT=0" npm install --legacy-peer-deps --no-optional

COPY webapp/ ./
RUN npm run pack

### 2. Build Go Server
FROM golang:1.21-bookworm AS gobuild

WORKDIR /go/src/focalboard
COPY . /go/src/focalboard

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev make && rm -rf /var/lib/apt/lists/*

RUN make templates-archive && EXCLUDE_PLUGIN=true EXCLUDE_SERVER=true EXCLUDE_ENTERPRISE=true make server-docker os=${TARGETOS} arch=${TARGETARCH}

### 3. Final Production Image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /opt/focalboard/data/files && \
    chown -R nobody:nogroup /opt/focalboard

WORKDIR /opt/focalboard

COPY --from=nodebuild --chown=nobody:nogroup /webapp/pack pack/
COPY --from=gobuild --chown=nobody:nogroup /go/src/focalboard/bin/docker/focalboard-server bin/
COPY --from=gobuild --chown=nobody:nogroup /go/src/focalboard/LICENSE.txt LICENSE.txt
COPY --from=gobuild --chown=nobody:nogroup /go/src/focalboard/docker/server_config.json config.json

USER nobody

EXPOSE 8000/tcp 9092/tcp

VOLUME /opt/focalboard/data

CMD ["/opt/focalboard/bin/focalboard-server"]

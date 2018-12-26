FROM golang:1.11.4 AS builder

# Take setted in docker compose hub revision
ARG HUB_REVISION

# Install golang dependencies manager
RUN go get -u github.com/golang/dep/cmd/dep && dep version

# Create work directory
WORKDIR $GOPATH/src/github.com/bitlum/hub/

# Download revision version of hub
RUN echo "HUB_REVISION: ${HUB_REVISION}" && \
    curl -L https://github.com/bitlum/hub/archive/$HUB_REVISION.tar.gz | tar xz --strip 1

# Install dependencies
RUN dep ensure -v

# Build hub
RUN go install . ./cmd/...

FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y ca-certificates curl
RUN update-ca-certificates
RUN rm -rf /var/lib/apt/lists/*

# This implies that service has been build of previos stage of dockerfile.
COPY --from=builder go/bin/hub /usr/local/bin/hub
COPY --from=builder go/bin/hubcli  /usr/local/bin/hubcli

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY hub.mainnet.conf /root/default/hub.conf

# Entrypoint script used to init datadir if required and for
# starting dash daemon
COPY entrypoint.sh /root/

ENTRYPOINT ["bash", "/root/entrypoint.sh"]

FROM golang:1.10.3 AS builder

# Take setted in docker compose hub revision
ARG HUB_REVISION

# Install golang dependencies manager
RUN go get -u github.com/golang/dep/cmd/dep

# Create work directory
WORKDIR $GOPATH/src/github.com/bitlum/hub/

# Download revision version of hub
RUN curl -L https://github.com/bitlum/hub/archive/$HUB_REVISION.tar.gz | tar xz --strip 1

# Go into direction where manager aka hub resides
WORKDIR $GOPATH/src/github.com/bitlum/hub/manager

# Install dependencies
RUN dep ensure -v

# Build hub
RUN go install

FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y ca-certificates curl
RUN update-ca-certificates
RUN rm -rf /var/lib/apt/lists/*

# This implies that service has to be built locally first, and putted
# in the docker directory, for running docker build.
COPY --from=builder /go/bin/manager /usr/local/bin/hub

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY hub.testnet.conf /root/default/hub.conf

# Entrypoint script used to init datadir if required and for
# starting dash daemon
COPY entrypoint.sh /root/

ENTRYPOINT ["bash", "/root/entrypoint.sh"]

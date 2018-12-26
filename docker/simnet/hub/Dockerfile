FROM ubuntu:16.04
MAINTAINER Samokhvalov Andrey <andrey@bitlum.io>

# Install root certificates so that we could fetch https links.
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates

# This implies that service has to be built locally first, and putted in the
# docker directory, fore running docker build.
COPY /bin/hub /usr/local/bin/
RUN chmod -R +x /usr/local/bin/hub

# This implies that service config is defines and located in the
# directory with docker file.
COPY hub.simnet.conf /root/default/hub.conf

# Entrypoint script used to init datadir if required and for
# starting dash daemon
COPY entrypoint.sh /root/

ENTRYPOINT [ "bash", "/root/entrypoint.sh" ]
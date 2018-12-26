#!/usr/bin/env bash

# This path is expected to be volume to make connector data persistent.
DATA_DIR=/root/.hub

# This path is expected to have default data used to init environment
# at first deploy such as config.
DEFAULTS_DIR=/root/default

CONFIG=${DATA_DIR}/hub.conf

# At first deploy datadir should not exists, we creating it.
if [[ ! -d ${DATA_DIR} ]]; then
    mkdir ${DATA_DIR}
fi

# We always restoring default config shipped with docker.
echo "Restoring default config"
cp ${DEFAULTS_DIR}/hub.conf ${CONFIG}

if [[ ${BITCOIN_RPC_USER} == "" ]]; then
    echo "WARN: Bitcoin rpc user is not specified"
    exit 1
fi

if [[ ${BITCOIN_RPC_PASSWORD} == "" ]]; then
    echo "WARN: Bitcoin rpc password is not specified"
    exit 1
fi

# We are using `exec` to enable gracefull shutdown of running daemon.
# Check http://veithen.github.io/2014/11/16/sigterm-propagation.html.
exec hub \
--config /root/.hub/hub.conf \
--bitcoind.user=${BITCOIN_RPC_USER} \
--bitcoind.pass=${BITCOIN_RPC_PASSWORD}
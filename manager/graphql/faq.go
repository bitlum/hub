package graphql

const faq = `# FAQ

# Q: What is hub?
# A: Hub is a lightning network node which has two primarial goals:
# 1. To be online 24 hours, 7 days in a week.
# 2. Have enough channel capacity to serve lightning
# network forwarding payments.

# Q: How to connect to hub lightning network node?
# A: Usually to connect to lightning network node you need to have
# [identityPubkey]@[host]:[port] and use it your wallet.
#
# Press "Run", and "GetHubInfo" to execute the
# query.
query GetHubInfo {
  info {
    network
    version

    neutrinoInfo {
      host
      port
    }

    nodeInfo {
      alias
      identityPubkey
      host
      port
    }

    blockHash
    blockHeight
  }
}

# Q: Could I see the activity in the local network of the hub?
# A: Yeah, you could see payments with obfuscated aliases.
#
# Press "Run", and "GetPayments" to execute the
# query.
query GetPayments {
  payments {
    fromPeer
    toPeer
    time
    amount
    paymentID
  }
}

# Q: How to get info about peers?
#
# Press "Run", and "GetPeers" to execute the
# query.
query GetPeers {
  peers {
    alias
    lockedByPeer
    lockedByHub
    isActive
    lastUpdate
  }
}
`

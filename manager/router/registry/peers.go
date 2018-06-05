package registry

import (
	"sync"
	"time"
	"github.com/bitlum/hub/manager/router"
	"math/rand"
	"strings"
)

// knownPeers this map is used to put peers which are initially publicly
// known, and which names might be reveal. Later when more payments are made,
// we populate it with peer random pseudonyms.
var knownPeers = make(map[router.UserID]string)
var knownPeersMutex sync.Mutex

// AddKnownPeer is used to add known peer's alias dynamically.
func AddKnownPeer(id router.UserID, alias string) {
	knownPeersMutex.Lock()
	defer knownPeersMutex.Unlock()
	knownPeers[id] = alias
}

// getRandomPseudonym returns random Sseudonym to obscure the real
// identification of receiver/sender.
func getRandomPseudonym() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return names[rand.Intn(len(names))]
}

// GetAlias return the alias by the given public key of the receiver/server,
// if node is not in the public list, than we obscure the name.
func GetAlias(userID router.UserID) string {
	knownPeersMutex.Lock()
	defer knownPeersMutex.Unlock()

	alias, ok := knownPeers[userID]
	if !ok {
		alias = getRandomPseudonym()
	}

	return strings.ToLower(alias)
}

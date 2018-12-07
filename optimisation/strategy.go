package optimisation

import "github.com/bitlum/hub/lightning"

// Defer is function which changes the state of the client, but executed a bit
// afterward.
type Defer func(client lightning.Client) error

// NodeStateStrategy represent the entity which takes the current client
// network state, desired one, and return set of actions which needed to apply
// to client to get it to optimum state.
type NodeStateStrategy interface {
	// TODO(andrew.shvv) Remove that because we switched to the UpdateChannel
	GenerateActions(oldState []*lightning.Channel,
		newState []*lightning.Channel) []Defer
}

// channelRebalance implements the NodeStateStrategy interface and represent
// the type of optimisation where we try to come to the equilibrium state by
// making the number of circular payments in the lightning network,
// so called off chain refinancing.
type channelRebalance struct{}

// channelUpdate implements NodeStateStrategy interface and repsents
type channelUpdate struct{}

// Runtime check that channelUpdate implements NodeStateStrategy interface.
var _ NodeStateStrategy = (*channelUpdate)(nil)

func NewChannelUpdateStrategy() *channelUpdate {
	return &channelUpdate{}
}

// GenerateActions generates the sequence of actions needed to bring the
// state to the optimisation equilibrium.
func (r *channelUpdate) GenerateActions(oldState []*lightning.Channel,
	newState []*lightning.Channel) []Defer {

	oldStateMap := listToMap(oldState)
	newStateMap := listToMap(newState)

	var actions []Defer
	for chanID, oldChan := range oldStateMap {
		chanIDCur := chanID
		if newChan, ok := newStateMap[chanID]; ok {
			// If new channel in the old channel map,
			// than the client balance has change.
			if newChan != oldChan {
				actions = append(actions, func(client lightning.Client) error {
					return client.UpdateChannel(chanIDCur, newChan.LocalBalance)
				})
			}
		} else {
			// If new channel not in the old channel map, that channel was
			// removed.
			actions = append(actions, func(client lightning.Client) error {
				return client.CloseChannel(chanIDCur)
			})
		}
	}

	for chanID, newChan := range newStateMap {
		if _, ok := oldStateMap[chanID]; !ok {
			// If old channel not in the new channel map, that channel was
			// added.
			actions = append(actions, func(client lightning.Client) error {
				return client.OpenChannel(newChan.UserID, newChan.LocalBalance)
			})
		}
	}

	return actions
}

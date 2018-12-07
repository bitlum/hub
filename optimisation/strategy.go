package optimisation

import "github.com/bitlum/hub/manager/router"

// Defer is function which changes the state of the router, but executed a bit
// afterward.
type Defer func(r router.Router) error

// RouterStateStrategy represent the entity which takes the current router
// network state, desired one, and return set of actions which needed to apply
// to router to get it to optimum state.
type RouterStateStrategy interface {
	// TODO(andrew.shvv) Remove that because we switched to the UpdateChannel
	GenerateActions(oldState []*router.Channel,
		newState []*router.Channel) []Defer
}

// channelRebalance implements the RouterStateStrategy interface and represent
// the type of optimisation where we try to come to the equilibrium state by
// making the number of circular payments in the lightning network,
// so called off chain refinancing.
type channelRebalance struct{}

// channelUpdate implements RouterStateStrategy interface and repsents
type channelUpdate struct{}

// Runtime check that channelUpdate implements RouterStateStrategy interface.
var _ RouterStateStrategy = (*channelUpdate)(nil)

func NewChannelUpdateStrategy() *channelUpdate {
	return &channelUpdate{}
}

// GenerateActions generates the sequence of actions needed to bring the
// state to the optimisation equilibrium.
func (r *channelUpdate) GenerateActions(oldState []*router.Channel,
	newState []*router.Channel) []Defer {

	oldStateMap := listToMap(oldState)
	newStateMap := listToMap(newState)

	var actions []Defer
	for chanID, oldChan := range oldStateMap {
		chanIDCur := chanID
		if newChan, ok := newStateMap[chanID]; ok {
			// If new channel in the old channel map,
			// than the router balance has change.
			if newChan != oldChan {
				actions = append(actions, func(r router.Router) error {
					return r.UpdateChannel(chanIDCur, newChan.RouterBalance)
				})
			}
		} else {
			// If new channel not in the old channel map, that channel was
			// removed.
			actions = append(actions, func(r router.Router) error {
				return r.CloseChannel(chanIDCur)
			})
		}
	}

	for chanID, newChan := range newStateMap {
		if _, ok := oldStateMap[chanID]; !ok {
			// If old channel not in the new channel map, that channel was
			// added.
			actions = append(actions, func(r router.Router) error {
				return r.OpenChannel(newChan.UserID, newChan.RouterBalance)
			})
		}
	}

	return actions
}

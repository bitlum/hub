package router

// Defer is function which changes the state of the router, but executed a bit
// afterward.
type Defer func(r Router) error

// RouterStateStrategy represent the entity which takes the current router
// network state, desired one, and return set of actions which needed to apply
// to router to get it to optimum state.
type RouterStateStrategy interface {
	GenerateActions(oldState []*Channel,
		newState []*Channel) []Defer
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
func (r *channelUpdate) GenerateActions(oldState []*Channel,
	newState []*Channel) []Defer {

	oldStateMap := listToMap(oldState)
	newStateMap := listToMap(newState)

	var actions []Defer
	for chanID, _ := range oldStateMap {
		if newChan, ok := newStateMap[chanID]; ok {
			// If new channel in the old channel map,
			// than the router balance has change.
			actions = append(actions, func(r Router) error {
				return r.UpdateChannel(chanID, newChan.RouterBalance)
			})
		} else {
			// If new channel not in the old channel map, that channel was
			// removed.
			actions = append(actions, func(r Router) error {
				return r.CloseChannel(chanID)
			})
		}
	}

	for chanID, newChan := range newStateMap {
		if _, ok := oldStateMap[chanID]; !ok {
			// If old channel not in the new channel map, that channel was
			// added.
			actions = append(actions, func(r Router) error {
				return r.OpenChannel(newChan.UserID, newChan.RouterBalance)
			})
		}
	}

	return actions
}

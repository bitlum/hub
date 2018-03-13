package router

// Defer...
type Defer func(r Router) error

// RebalancingStrategy represent the entity which takes the current router
// network state, desired one, and return set of actions which needed to apply
// to router to get it to optimum state.
type RebalancingStrategy interface {
	GenerateActions(oldState []*Channel,
		newState []*Channel) []Defer
}

// channelRebalance implements the RebalancingStrategy interface and represent
// the type of optimisation where we try to come to the equilibrium state by
// making the number of circular payments in the lightning network,
// so called off chain refinancing.
type channelRebalance struct{}

// channelRecreation implements RebalancingStrategy interface and repsent
type channelRecreation struct{}

// Runtime check that channelRecreation implements RebalancingStrategy interface.
var _ RebalancingStrategy = (*channelRecreation)(nil)

func NewRebalancingStrategy() *channelRecreation {
	return &channelRecreation{}
}

func (r *channelRecreation) GenerateActions(oldState []*Channel,
	newState []*Channel) []Defer {
	// TODO(andrew.shhvv) Find difference between old and new
	// TODO(andrew.shhvv) Generate actions of open/close channels

	// Do nothing and exit
	return nil
}

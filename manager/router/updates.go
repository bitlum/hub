package router

import "time"

type UpdateChannelClosing struct {
	UserID    UserID
	ChannelID ChannelID

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

type UpdateChannelClosed struct {
	UserID    UserID
	ChannelID ChannelID

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration time.Duration
}

// UpdateChannelUpdating is used to notify that one of the participants
// decided to splice in or splice out some portion of their money from the
// channel.
//
// NOTE: On 11.03.2018 this is not yet possible in the Bitcoin Lightning
// Network, channel might be either opened or closed.
type UpdateChannelUpdating struct {
	UserID    UserID
	ChannelID ChannelID

	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

// UpdateChannelUpdated is used to notify that one of the participants
// decided to splice in or splice out some portion of their money from the
// channel.
//
// NOTE: On 11.03.2018 this is not yet possible in the Bitcoin Lightning
// Network, channel might be either opened or closed.
type UpdateChannelUpdated struct {
	UserID    UserID
	ChannelID ChannelID

	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration time.Duration
}

// UpdateChannelOpening is used as notifications from router or network that
// channel started to opening, and wait for blockchain confirmation.
type UpdateChannelOpening struct {
	UserID        UserID
	ChannelID     ChannelID
	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

// UpdateChannelOpened is used as notifications from router or network that
// channel has been opened.
type UpdateChannelOpened struct {
	UserID    UserID
	ChannelID ChannelID

	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration time.Duration
}

type UpdatePayment struct {
	Status PaymentStatus
	Type   PaymentType

	Sender   UserID
	Receiver UserID

	Amount BalanceUnit

	// Earned is the number of funds which router earned by making this payment.
	// In case of re-balancing router will pay the fee, for that reason this
	// number will be negative.
	Earned BalanceUnit
}

// UpdateLinkAverageUpdateDuration is used when router wants to notify that
// the average link update time has changed.
type UpdateLinkAverageUpdateDuration struct {
	AverageUpdateDuration time.Duration
}

// UpdateUserActive notify that user with given id is online or offline,
// which means that all associated with him channels could or couldn't be used
// for forwarding payments.
type UpdateUserActive struct {
	User     UserID
	IsActive bool
}

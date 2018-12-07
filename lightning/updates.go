package lightning

type UpdateChannelClosing struct {
	UserID    UserID
	ChannelID ChannelID

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

func (u *UpdateChannelClosing) String() string {
	return "channel_closing"
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
	Duration int64
}

func (u *UpdateChannelClosed) String() string {
	return "channel_closed"
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

	RemoteBalance BalanceUnit
	LocalBalance  BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

func (u *UpdateChannelUpdating) String() string {
	return "channel_updating"
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

	RemoteBalance BalanceUnit
	LocalBalance  BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration int64
}

func (u *UpdateChannelUpdated) String() string {
	return "channel_updated"
}

// UpdateChannelOpening is used as notifications from lightning client or
// network that channel started to opening, and wait for blockchain confirmation.
type UpdateChannelOpening struct {
	UserID        UserID
	ChannelID     ChannelID
	RemoteBalance BalanceUnit
	LocalBalance  BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

func (u *UpdateChannelOpening) String() string {
	return "channel_opening"
}

// UpdateChannelOpened is used as notifications from lightning client or network
// that channel has been opened.
type UpdateChannelOpened struct {
	UserID    UserID
	ChannelID ChannelID

	RemoteBalance BalanceUnit
	LocalBalance  BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration int64
}

func (u *UpdateChannelOpened) String() string {
	return "channel_opened"
}

type UpdatePayment struct {
	ID string

	Status PaymentStatus
	Type   PaymentType

	Sender   UserID
	Receiver UserID

	Amount BalanceUnit

	// Earned is the number of funds which lightning node earned by making
	// this payment. In case of re-balancing lightning node will pay the fee,
	// for that
	// reason this number will be negative.
	Earned BalanceUnit
}

func (u *UpdatePayment) String() string {
	return "payment"
}

// UpdateUserConnected notify that user with given id is online or offline,
// which means that all associated with him channels could or couldn't be used
// for forwarding payments.
type UpdateUserConnected struct {
	User        UserID
	IsConnected bool
}

func (u *UpdateUserConnected) String() string {
	return "user_connected/disconnected"
}

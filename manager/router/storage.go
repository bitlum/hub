package router

// ChannelStorage contains all necessary hooks to store channel object.
type ChannelStorage interface {
	// AddChannelState adds state to the channel's state array. State array
	// should be initialised in the Channel object on the stage of getting
	// channels.
	AddChannelState(chanID ChannelID, state *ChannelState) error

	// UpdateChannel saves or update channel without saving its states.
	UpdateChannel(channel *Channel) error

	// Channels returns all channels and its states.
	Channels() ([]*Channel, error)

	// RemoveChannel removes the channel.
	RemoveChannel(channel *Channel) error
}

// UserStorage contains all necessary hooks to store user object.
type UserStorage interface {
	// UpdateUser creates or updates user in database.
	UpdateUser(user *User) error

	// Users returns all users who are somehow related or were realted to the
	// hub.
	Users() ([]*User, error)
}

// InfoStorage contains all necessary hooks to store hub info objects.
type InfoStorage interface {
	// UpdateInfo updates information about the hub lightning network node.
	UpdateInfo(info *Info) error

	// Info returns hub lighting network node information.
	Info() (*Info, error)
}

// PaymentStorage contains all necessary hooks to store payment object.
type PaymentStorage interface {
	// Payments returns the payments happening inside the hub local network,
	Payments() ([]*Payment, error)

	// StorePayment saves the payment.
	StorePayment(payment *Payment) error
}

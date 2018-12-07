package router

import (
	"github.com/go-errors/errors"
)

// UserID uniquely identifies the user in the local lightning network.
type UserID string

type UserConfig struct {
	// Storage is used to keep important data persistent.
	Storage UserStorage
}

func (c *UserConfig) validate() error {
	if c == nil {
		return errors.Errorf("config is nil")
	}

	if c.Storage == nil {
		return errors.Errorf("storage is empty")
	}

	return nil
}

type User struct {
	UserID UserID

	// IsConnected is user connected to hub with tcp/ip connection and active
	// for communication.
	IsConnected bool

	Alias        string
	LockedByHub  BalanceUnit
	LockedByUser BalanceUnit

	cfg *UserConfig

	// TODO(andrew.shvv) make user store channels
}

func NewUser(userID UserID, isConnected bool, alias string,
	lockedByHub, lockedByUser BalanceUnit, cfg *UserConfig) (*User, error) {

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &User{
		UserID:       userID,
		IsConnected:  isConnected,
		Alias:        alias,
		LockedByHub:  lockedByHub,
		LockedByUser: lockedByUser,
		cfg:          &(*cfg),
	}, nil
}

// Save creates or updates user in database.
func (u *User) Save() error {
	return u.cfg.Storage.UpdateUser(u)
}

// SetConfig is used to set config which is used for using the external
// subsystems.
func (u *User) SetConfig(cfg *UserConfig) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	// Copy config file
	u.cfg = &(*cfg)
	return nil
}

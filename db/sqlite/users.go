package sqlite

import "github.com/bitlum/hub/lightning"

// Runtime check to ensure that DB implements lightning.UserStorage interface.
var _ lightning.UserStorage = (*DB)(nil)

// UpdateUser updates information about set of online and peers
// connected to the hub.
func (d *DB) UpdateUser(user *lightning.User) error {
	return d.Save(&User{
		ID:           string(user.UserID),
		Alias:        user.Alias,
		IsConnected:  user.IsConnected,
		LockedByUser: int64(user.LockedByUser),
		LockedByHub:  int64(user.LockedByHub),
	}).Error
}

// Users loads all users which are related to the hub.
func (d *DB) Users() ([]*lightning.User, error) {
	var users []User
	db := d.Find(&users)
	if err := db.Error; err != nil {
		return nil, err
	}

	prs := make([]*lightning.User, len(users))
	for i, user := range users {
		prs[i] = &lightning.User{
			UserID:       lightning.UserID(user.ID),
			IsConnected:  user.IsConnected,
			Alias:        user.Alias,
			LockedByUser: lightning.BalanceUnit(user.LockedByUser),
			LockedByHub:  lightning.BalanceUnit(user.LockedByHub),
		}
	}

	return prs, nil
}

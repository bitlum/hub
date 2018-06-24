package sqlite

import "github.com/bitlum/hub/manager/router"

// Runtime check to ensure that DB implements router.UserStorage interface.
var _ router.UserStorage = (*DB)(nil)

// UpdateUser updates information about set of online and peers
// connected to the hub.
func (d *DB) UpdateUser(user *router.User) error {
	return d.Save(&User{
		ID:           string(user.UserID),
		Alias:        user.Alias,
		IsConnected:  user.IsConnected,
		LockedByUser: int64(user.LockedByUser),
		LockedByHub:  int64(user.LockedByHub),
	}).Error
}

// Users loads all users which are related to the hub.
func (d *DB) Users() ([]*router.User, error) {
	var users []User
	db := d.Find(&users)
	if err := db.Error; err != nil {
		return nil, err
	}

	prs := make([]*router.User, len(users))
	for i, user := range users {
		prs[i] = &router.User{
			UserID:       router.UserID(user.ID),
			IsConnected:  user.IsConnected,
			Alias:        user.Alias,
			LockedByUser: router.BalanceUnit(user.LockedByUser),
			LockedByHub:  router.BalanceUnit(user.LockedByHub),
		}
	}

	return prs, nil
}

package sqlite

import (
	"github.com/bitlum/hub/lightning"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"os"
	"path/filepath"
)

// DB is the primary datastore.
type DB struct {
	*gorm.DB
	dbPath string

	// nodeInfo is a information about hub, which is stored in-memory.
	nodeInfo *lightning.Info
}

// Open opens an existing db. Any necessary schemas migrations due to
// updates will take place as necessary.
func Open(dbPath string, dbName string) (*DB, error) {
	path := filepath.Join(dbPath, dbName)

	if !fileExists(dbPath) {
		if err := os.MkdirAll(dbPath, 0700); err != nil {
			return nil, err
		}
	}

	gdb, err := gorm.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = gdb.AutoMigrate(
		&Counters{},
		&Channel{},
		&Payment{},
		&User{},
		&State{},
		&ChannelIDShortChanIDIndex{},
		&UserIDShortChanIDIndex{}).Error
	if err != nil {
		return nil, err
	}

	db := &DB{
		DB:     gdb,
		dbPath: dbPath,
	}

	return db, nil
}

// fileExists returns true if the file exists, and false otherwise.
func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

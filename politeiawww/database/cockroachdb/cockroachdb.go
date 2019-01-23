package cockroachdb

import (
	"encoding/base64"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/decred/politeia/politeiawww/database"

	"github.com/jinzhu/gorm"
)

const (
	dbPrefix = "users_"

	// UserPoliteiawww is a database user with read/write access
	UserPoliteiawww = "politeiawww"

	// Database table names
	tableUsers   = "users"
	tableVersion = "version"

	// UserVersion is the curent database version
	UserVersion uint32 = 1
)

// Cockroachdb implements the database interface
type cockroachdb struct {
	sync.RWMutex
	shutdown bool     // Backend is shutdown
	usersdb  *gorm.DB // Database context
}

// Set is used to store a user payload by a given user ID
func (c *cockroachdb) Set(userID string, payload []byte) error {
	log.Tracef("Set: %v", userID)

	c.RLock()
	shutdown := c.shutdown
	c.RUnlock()

	if shutdown {
		return database.ErrShutdown
	}

	// run Set within a transaction
	tx := c.usersdb.Begin()

	// try to find the user with the given user ID
	var user *User
	err := tx.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// convert user payload from bytes to string
	stringEncodedPayload := base64.StdEncoding.EncodeToString(payload)

	if user != nil {
		// update existent user
		err = tx.Model(&user).Update(&User{
			Payload: stringEncodedPayload,
		}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// create a new user
		err = tx.Create(&User{
			UserID:  userID,
			Payload: stringEncodedPayload,
		}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Get returns the user payload by a given user ID
func (c *cockroachdb) Get(userID string) ([]byte, error) {
	log.Tracef("Get: %v", userID)

	c.RLock()
	shutdown := c.shutdown
	c.RUnlock()

	if shutdown {
		return nil, database.ErrShutdown
	}

	// find user by id
	var user *User
	err := c.usersdb.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(user.Payload)
}

func createTables(db *gorm.DB) error {
	log.Tracef("createTables")
	if !db.HasTable(tableVersion) {
		err := db.CreateTable(&Version{}).Error
		if err != nil {
			return err
		}

		err = db.Create(&Version{
			Version:   UserVersion,
			Timestamp: time.Now().Unix(),
		}).Error
		if err != nil {
			return err
		}
	}

	if !db.HasTable(tableUsers) {
		err := db.CreateTable(&User{}).Error
		if err != nil {
			return err
		}
	}

	return nil

}

// Create uses the CockroachDB root user to create a database,
// politeiawww user if it does not already exist. User permissions are then
// set for the database and the database tables are created if they do not
// already exist. A Version record is inserted into the database during table
// creation.
func Create(host, net, rootCert, certDir string) error {
	log.Tracef("Create: %v %v %v %v", host, net, rootCert, certDir)

	// Connect to CockroachDB as root user. CockroachDB connects
	// to defaultdb when a database is not specified.
	h := "postgresql://root@" + host
	u, err := url.Parse(h)
	if err != nil {
		log.Debugf("Create: could not parse url %v", h)
		return err
	}

	v := url.Values{}
	v.Set("ssl", "true")
	v.Set("sslmode", "require")
	v.Set("sslrootcert", filepath.Clean(rootCert))
	v.Set("sslkey", filepath.Join(certDir, "client."+u.User.String()+".key"))
	v.Set("sslcert", filepath.Join(certDir, "client."+u.User.String()+".crt"))

	addr := u.String() + "?" + v.Encode()
	db, err := gorm.Open("postgres", addr)
	defer db.Close()
	if err != nil {
		log.Debugf("Create: could not connect to %v", addr)
		return err
	}

	// Setup politeiawww database and users
	dbName := dbPrefix + net
	q := "CREATE DATABASE IF NOT EXISTS " + dbName
	err = db.Exec(q).Error
	if err != nil {
		return err
	}

	q = "CREATE USER IF NOT EXISTS " + UserPoliteiawww
	err = db.Exec(q).Error
	if err != nil {
		return err
	}
	q = "GRANT ALL ON DATABASE " + dbName + " TO " + UserPoliteiawww
	err = db.Exec(q).Error
	if err != nil {
		return err
	}

	// Connect to records database with root user
	h = "postgresql://root@" + host + "/" + dbName
	u, err = url.Parse(h)
	if err != nil {
		log.Debugf("Create: could not parse url %v", h)
		return err
	}
	addr = u.String() + "?" + v.Encode()
	pdb, err := gorm.Open("postgres", addr)
	defer pdb.Close()
	if err != nil {
		log.Debugf("Create: could not connect to %v", addr)
		return err
	}

	// Setup database tables
	tx := pdb.Begin()
	err = createTables(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

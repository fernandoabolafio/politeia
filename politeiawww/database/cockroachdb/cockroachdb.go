package cockroachdb

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/marcopeereboom/sbox"

	"github.com/decred/politeia/politeiawww/database"
	"github.com/jinzhu/gorm"
)

const (
	dbPrefix = "users_"

	// UserPoliteiawww is a database user with read/write access
	UserPoliteiawww = "politeiawww"

	// Database table names
	tableKeyValue = "key_value"

	// UserVersion is the curent database version
	UserVersion uint32 = 1
)

var (
	_ database.Database = (*cockroachdb)(nil)
)

// cockroachdb implements the database interface
type cockroachdb struct {
	sync.RWMutex
	shutdown      bool                    // Backend is shutdown
	usersdb       *gorm.DB                // Database context
	encryptionKey *database.EncryptionKey // Encryption key
	dbAddress     string                  // Database address
}

func buildDbQueryString(rootCert, certDir string, u *url.URL) string {
	v := url.Values{}
	v.Set("ssl", "true")
	v.Set("sslmode", "require")
	v.Set("sslrootcert", filepath.Clean(rootCert))
	v.Set("sslkey", filepath.Join(certDir, "client."+u.User.String()+".key"))
	v.Set("sslcert", filepath.Join(certDir, "client."+u.User.String()+".crt"))

	return v.Encode()
}

// Put stores a payload by a given key
func (c *cockroachdb) Put(key string, payload []byte) error {
	log.Tracef("Put: %v", key)

	c.RLock()
	shutdown := c.shutdown
	c.RUnlock()

	if shutdown {
		return database.ErrShutdown
	}

	// run Put within a transaction
	tx := c.usersdb.Begin()

	// encrypt payload
	packed, err := sbox.Encrypt(database.DatabaseVersion, &c.encryptionKey.Key, payload)
	if err != nil {
		return err
	}

	// try to find the record with the provided key
	var keyValue KeyValue
	err = tx.Where("key = ?", key).First(&keyValue).Error
	if gorm.IsRecordNotFoundError(err) {
		// record not found, so we creaet a new one
		err = tx.Create(&KeyValue{
			Key:     key,
			Payload: packed,
		}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	} else if err != nil {
		// return any other error
		tx.Rollback()
		return err
	} else {
		// record found, update existent value
		err = tx.Model(&keyValue).Update(&KeyValue{
			Payload: packed,
		}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Get returns a payload by a given key
func (c *cockroachdb) Get(key string) ([]byte, error) {
	log.Tracef("Get: %v", key)

	c.RLock()
	shutdown := c.shutdown
	c.RUnlock()

	if shutdown {
		return nil, database.ErrShutdown
	}

	// find user by id
	var keyValue KeyValue
	err := c.usersdb.Where("key = ?", key).First(&keyValue).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	payload, _, err := sbox.Decrypt(&c.encryptionKey.Key, keyValue.Payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *cockroachdb) GetAll(callbackFn func(string, []byte)) error {
	log.Tracef("GetAll")

	c.RLock()
	shutdown := c.shutdown
	c.RUnlock()

	if shutdown {
		return database.ErrShutdown
	}

	var values []KeyValue
	err := c.usersdb.Find(&values).Error
	if err != nil {
		return err
	}
	for _, v := range values {
		// decrypt payload
		decValue, _, err := sbox.Decrypt(&c.encryptionKey.Key, v.Payload)
		if err != nil {
			return err
		}
		// fmt.Printf("KEY: %v, VALUE: ")
		callbackFn(v.Key, decValue)
	}

	return nil
}

// Has returns true if the database does contains the given key.
func (c *cockroachdb) Has(key string) (bool, error) {
	log.Tracef("Has: %v", key)

	c.RLock()
	shutdown := c.shutdown
	c.RUnlock()

	if shutdown {
		return false, database.ErrShutdown
	}

	var keyValue KeyValue
	err := c.usersdb.Where("key = ?", key).First(&keyValue).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil

}

// Close shuts down the database.  All interface functions MUST return with
// errShutdown if the backend is shutting down.
func (c *cockroachdb) Close() error {
	log.Tracef("Close")

	c.Lock()
	defer c.Unlock()

	c.shutdown = true
	return c.usersdb.Close()
}

func createTables(db *gorm.DB) error {
	log.Tracef("createTables")

	if !db.HasTable(tableKeyValue) {
		err := db.CreateTable(&KeyValue{}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateCDB uses the CockroachDB root user to create a database,
// politeiawww user if it does not already exist. User permissions are then
// set for the database and the database tables are created if they do not
// already exist. The encryption key is also created in case it does not exist.
func CreateCDB(host, net, rootCert, certDir, keyDir string) error {
	log.Tracef("Create: %v %v %v %v", host, net, rootCert, certDir)

	// Connect to CockroachDB as root user. CockroachDB connects
	// to defaultdb when a database is not specified.
	h := "postgresql://root@" + host
	u, err := url.Parse(h)
	if err != nil {
		log.Debugf("Create: could not parse url %v", h)
		return err
	}

	qs := buildDbQueryString(rootCert, certDir, u)

	addr := u.String() + "?" + qs

	fmt.Print(addr)

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
	addr = u.String() + "?" + qs
	pdb, err := gorm.Open("postgres", addr)
	defer pdb.Close()
	if err != nil {
		log.Debugf("Create: could not connect to %v", addr)
		return err
	}

	// see if we need to create a new encryption key
	err = database.ResolveEncryptionKey(keyDir)
	if err != nil {
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

// Open opens a new database connection and make sure there is a version record
// stored in the database
func (c *cockroachdb) Open() error {
	//open a new database connection
	db, err := gorm.Open("postgres", c.dbAddress)
	if err != nil {
		log.Debugf("Open: could not connect to %v", c.dbAddress)
		return err
	}

	c.usersdb = db

	// see if we need to write a version record
	payload, err := c.Get(database.DatabaseVersionKey)

	if err == database.ErrNotFound {
		// write version record
		payload, err = database.EncodeVersion(database.Version{
			Version: database.DatabaseVersion,
			Time:    time.Now().Unix(),
		})
		if err != nil {
			return err
		}
		fmt.Printf("got here")
		return c.Put(database.DatabaseVersionKey, payload)
	}

	if err != nil {
		return err
	}

	return nil
}

// NewCDB returns a new cockroachdb context that contains a connection to the
// specified database that was made using the passed in user and certificates.
func NewCDB(user, host, net, rootCert, certDir, keyDir string) (*cockroachdb, error) {
	log.Tracef("New: %v %v %v %v %v", user, host, net, rootCert, certDir)

	// Connect to database
	h := "postgresql://" + user + "@" + host + "/" + dbPrefix + net
	u, err := url.Parse(h)
	if err != nil {
		log.Debugf("New: could not parse url %v", h)
		return nil, err
	}

	qs := buildDbQueryString(rootCert, certDir, u)

	addr := u.String() + "?" + qs

	// load encryption key
	ek, err := database.LoadEncryptionKey(filepath.Join(keyDir, database.DefaultEncryptionKeyFilename))
	if err != nil {
		fmt.Printf("error %v", err)
		return nil, database.ErrLoadingEncryptionKey
	}

	c := &cockroachdb{
		dbAddress:     addr,
		encryptionKey: ek,
	}

	// Open the database
	err = c.Open()
	if err != nil {
		return nil, err
	}

	// Disable gorm logging. This prevents duplicate errors from
	// being printed since we handle errors manually.
	c.usersdb.LogMode(false)

	// Disable automatic table name pluralization. We set table
	// names manually.
	c.usersdb.SingularTable(true)

	return c, err
}

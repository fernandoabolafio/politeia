// Copyright (c) 2017-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package leveldb

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/decred/politeia/politeiawww/database"
	ldb "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	UserdbPath = "users"
)

var (
	_ database.Database = (*leveldb)(nil)
)

// leveldb implements the database interface.
type leveldb struct {
	sync.RWMutex
	shutdown      bool                    // Backend is shutdown
	root          string                  // Database root
	userdb        *ldb.DB                 // Database context
	encryptionKey *database.EncryptionKey // Encryption key
}

// Put stores a payload by a given key
func (l *leveldb) Put(key string, payload []byte) error {
	log.Tracef("Put %v:", key)

	l.RLock()
	shutdown := l.shutdown
	l.RUnlock()

	if shutdown {
		return database.ErrShutdown
	}

	// encrypt payload
	packed, err := database.Encrypt(database.DatabaseVersion, l.encryptionKey.Key, payload)
	if err != nil {
		return err
	}

	return l.userdb.Put([]byte(key), packed, nil)
}

// Get returns a payload by a given key
func (l *leveldb) Get(key string) ([]byte, error) {
	log.Tracef("Get: %v", key)

	l.RLock()
	shutdown := l.shutdown
	l.RUnlock()

	if shutdown {
		return nil, database.ErrShutdown
	}

	packed, err := l.userdb.Get([]byte(key), nil)
	if err == ldb.ErrNotFound {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	payload, _, err := database.Decrypt(l.encryptionKey.Key, packed)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (l *leveldb) GetAll(callbackFn func(string, []byte)) error {
	l.RLock()
	shutdown := l.shutdown
	l.RUnlock()

	if shutdown {
		return database.ErrShutdown
	}

	iter := l.userdb.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		// decrypt value
		decValue, _, err := database.Decrypt(l.encryptionKey.Key, value)
		if err != nil {
			return err
		}

		callbackFn(string(key), decValue)
	}
	iter.Release()

	return iter.Error()
}

// Has returns true if the database does contains the given key.
func (l *leveldb) Has(key string) (bool, error) {
	l.RLock()
	shutdown := l.shutdown
	l.RUnlock()

	if shutdown {
		return false, database.ErrShutdown
	}

	return l.userdb.Has([]byte(key), nil)
}

// Open opens a new database connection and make sure there is a version record
// stored in the database
func (l *leveldb) Open() error {
	log.Tracef("Open leveldb")

	// open database
	var err error
	l.userdb, err = ldb.OpenFile(filepath.Join(l.root, UserdbPath), &opt.Options{
		ErrorIfMissing: true,
	})
	if err != nil {
		return err
	}

	// See if we need to write a version record
	payload, err := l.Get(database.DatabaseVersionKey)

	if err == database.ErrNotFound {
		// Write version record
		payload, err = database.EncodeVersion(database.Version{
			Version: database.DatabaseVersion,
			Time:    time.Now().Unix(),
		})
		if err != nil {
			return err
		}

		packed, err := database.Encrypt(database.DatabaseVersion,
			l.encryptionKey.Key, payload)
		if err != nil {
			return err
		}

		return l.Put(database.DatabaseVersionKey, packed)
	} else {
		// Version record already exists, so we check if the encryption key
		// is valid
		_, version, err := database.Decrypt(l.encryptionKey.Key, payload)
		if err != nil {
			return database.ErrWrongEncryptionKey
		}
		// Also check if the record version matches the interface implementation
		// version
		if version != database.DatabaseVersion {
			return database.ErrWrongVersion
		}
	}

	return err
}

// Close shuts down the database.  All interface functions MUST return with
// errShutdown if the backend is shutting down.
//
// Close satisfies the backend interface.
func (l *leveldb) Close() error {
	l.Lock()
	defer l.Unlock()

	l.shutdown = true
	return l.userdb.Close()
}

// CreateLevelDB creates a new leveldb database if does not already exist.
func CreateLevelDB(dataDir string) error {
	log.Tracef("Create LevelDB: %v %v", dataDir)

	// db openFile is called to make sure the db will be created in case it
	// doesn not exist
	db, err := ldb.OpenFile(filepath.Join(dataDir, UserdbPath), nil)
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	return nil
}

// NewLevelDB creates a new leveldb instance. It must be called after the Create
// method, otherwise it will throw an error.
func NewLevelDB(dataDir string, dbKey *database.EncryptionKey) (*leveldb, error) {
	log.Tracef("New LevelDB: %v %v", dataDir, dbKey)

	l := &leveldb{
		root:          dataDir,
		encryptionKey: dbKey,
	}

	err := l.Open()
	if err != nil {
		return nil, err
	}

	return l, nil
}

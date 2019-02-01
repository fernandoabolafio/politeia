package database

// import (
// 	"github.com/decred/politeia/politeiawww/database"
// )

// import (
// 	"encoding/binary"
// 	"strings"

// 	"github.com/badoux/checkmail"
// 	"github.com/google/uuid"
// 	"github.com/syndtr/goleveldb/leveldb"
// )

// // Store new user.
// //
// // UserNew satisfies the backend interface.
// func (l *localdb) UserNew(u database.User) error {
// 	return nil
// }

// // UserGet returns a user record if found in the database.
// //
// // UserGet satisfies the backend interface.
// func (l *localdb) UserGetByEmail(email string) (*database.User, error) {
// 	return nil, nil
// }

// // UserGetByUsername returns a user record given its username, if found in the database.
// //
// // UserGetByUsername satisfies the backend interface.
// func (l *localdb) UserGetByUsername(username string) (*database.User, error) {
// 	return nil, nil
// }

// // UserGetById returns a user record given its id, if found in the database.
// //
// // UserGetById satisfies the backend interface.
// func (l *localdb) UserGetByID(id uuid.UUID) (*database.User, error) {
// 	return nil, nil
// }

// // Update existing user.
// //
// // UserUpdate satisfies the backend interface.
// func (l *localdb) UserUpdate(u database.User) error {
// 	return nil
// }

// Update existing user.
//
// UserUpdate satisfies the backend interface.
// func (l *localdb) AllUsers(callbackFn func(u *database.User)) error {
// 	return nil
// }

// func (l *localdb) UserNew(u database.User) error {
// 	l.Lock()
// 	defer l.Unlock()

// 	if l.shutdown {
// 		return database.ErrShutdown
// 	}

// 	log.Debugf("UserNew: %v", u)

// 	if err := checkmail.ValidateFormat(u.Email); err != nil {
// 		return database.ErrInvalidEmail
// 	}

// 	// Make sure user does not exist
// 	ok, err := l.userdb.Has([]byte(u.Email), nil)
// 	if err != nil {
// 		return err
// 	} else if ok {
// 		return database.ErrUserExists
// 	}

// 	// Fetch the next unique paywall index for the user.
// 	var lastPaywallIndex uint64
// 	b, err := l.userdb.Get([]byte(LastPaywallAddressIndex), nil)
// 	if err != nil {
// 		if err != leveldb.ErrNotFound {
// 			return err
// 		}
// 	} else {
// 		lastPaywallIndex = binary.LittleEndian.Uint64(b) + 1
// 	}

// 	// Set the new paywall index on the user.
// 	u.PaywallAddressIndex = lastPaywallIndex

// 	// Write the new paywall index back to the db.
// 	b = make([]byte, 8)
// 	binary.LittleEndian.PutUint64(b, lastPaywallIndex)
// 	err = l.userdb.Put([]byte(LastPaywallAddressIndex), b, nil)
// 	if err != nil {
// 		return err
// 	}

// 	// Set unique uuid for the user.
// 	u.ID = uuid.New()

// 	payload, err := EncodeUser(u)
// 	if err != nil {
// 		return err
// 	}

// 	return l.userdb.Put([]byte(u.Email), payload, nil)
// }

// // UserGet returns a user record if found in the database.
// //
// // UserGet satisfies the backend interface.
// func (l *localdb) UserGetByEmail(email string) (*database.User, error) {
// 	l.Lock()
// 	defer l.Unlock()

// 	if l.shutdown {
// 		return nil, database.ErrShutdown
// 	}

// 	payload, err := l.userdb.Get([]byte(strings.ToLower(email)), nil)
// 	if err == leveldb.ErrNotFound {
// 		return nil, database.ErrNotFound
// 	} else if err != nil {
// 		return nil, err
// 	}

// 	u, err := DecodeUser(payload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return u, nil
// }

// // UserGetByUsername returns a user record given its username, if found in the database.
// //
// UserGetByUsername satisfies the backend interface.
// func (l *localdb) UserGetByUsername(username string) (*database.User, error) {
// 	l.Lock()
// 	defer l.Unlock()

// 	if l.shutdown {
// 		return nil, database.ErrShutdown
// 	}

// 	log.Debugf("UserGetByUsername\n")

// 	iter := l.userdb.NewIterator(nil, nil)
// 	for iter.Next() {
// 		key := iter.Key()
// 		value := iter.Value()

// 		if !isUserRecord(string(key)) {
// 			continue
// 		}

// 		user, err := DecodeUser(value)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if strings.ToLower(user.Username) == strings.ToLower(username) {
// 			return user, err
// 		}
// 	}
// 	iter.Release()

// 	return nil, iter.Error()
// }

// // UserGetById returns a user record given its id, if found in the database.
// //
// // UserGetById satisfies the backend interface.
// func (l *localdb) UserGetByID(id uuid.UUID) (*database.User, error) {
// 	l.Lock()
// 	defer l.Unlock()

// 	if l.shutdown {
// 		return nil, database.ErrShutdown
// 	}

// 	log.Debugf("UserGetById\n")

// 	iter := l.userdb.NewIterator(nil, nil)
// 	for iter.Next() {
// 		key := iter.Key()
// 		value := iter.Value()

// 		if !isUserRecord(string(key)) {
// 			continue
// 		}

// 		user, err := DecodeUser(value)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if user.ID == id {
// 			return user, err
// 		}
// 	}
// 	iter.Release()

// 	return nil, iter.Error()
// }

// // Update existing user.
// //
// // UserUpdate satisfies the backend interface.
// func (l *localdb) UserUpdate(u database.User) error {
// 	l.Lock()
// 	defer l.Unlock()

// 	if l.shutdown {
// 		return database.ErrShutdown
// 	}

// 	log.Debugf("UserUpdate: %v", u)

// 	// Make sure user already exists
// 	exists, err := l.userdb.Has([]byte(u.Email), nil)
// 	if err != nil {
// 		return err
// 	} else if !exists {
// 		return database.ErrNotFound
// 	}

// 	payload, err := EncodeUser(u)
// 	if err != nil {
// 		return err
// 	}

// 	return l.userdb.Put([]byte(u.Email), payload, nil)
// }

// // Update existing user.
// //
// // UserUpdate satisfies the backend interface.
// func (l *localdb) AllUsers(callbackFn func(u *database.User)) error {
// 	l.Lock()
// 	defer l.Unlock()

// 	if l.shutdown {
// 		return database.ErrShutdown
// 	}

// 	log.Debugf("AllUsers\n")

// 	iter := l.userdb.NewIterator(nil, nil)
// 	for iter.Next() {
// 		key := iter.Key()
// 		value := iter.Value()

// 		if !isUserRecord(string(key)) {
// 			continue
// 		}

// 		u, err := DecodeUser(value)
// 		if err != nil {
// 			return err
// 		}

// 		callbackFn(u)
// 	}
// 	iter.Release()

// 	return iter.Error()
// }

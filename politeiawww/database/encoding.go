package database

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/decred/politeia/util"
	"github.com/marcopeereboom/sbox"
)

func verifyRecordVersion(recordVersion, dbVersion uint32) error {
	if recordVersion != dbVersion {
		return ErrWrongRecordVersion
	}
	return nil
}

func verifyRecordType(recordType, expectedType RecordTypeT) error {
	if recordType != expectedType {
		return ErrWrongRecordType
	}
	return nil
}

// EncodeVersion encodes Version into a JSON byte slice. It also adds the
// record type and version before encoding.
func EncodeVersion(version Version) ([]byte, error) {
	// make sure it has record type and version specified
	version.RecordType = RecordTypeVersion
	version.RecordVersion = DatabaseVersion

	b, err := json.Marshal(version)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// DecodeVersion decodes a JSON byte slice into a Version.
func DecodeVersion(payload []byte) (*Version, error) {
	var version Version

	err := json.Unmarshal(payload, &version)
	if err != nil {
		return nil, err
	}

	err = verifyRecordVersion(version.RecordVersion, DatabaseVersion)
	if err != nil {
		return nil, err
	}

	err = verifyRecordType(version.RecordType, RecordTypeVersion)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

// EncodeUser encodes User into a JSON byte slice. It also adds the
// record type and record version before encoding.
func EncodeUser(u User) ([]byte, error) {
	// make sure it user has record type and version specified
	u.RecordType = RecordTypeUser
	u.RecordVersion = DatabaseVersion

	b, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// EncodeLastPaywallAddressIndex encodes User into a JSON byte slice.
// It also adds the record type and version before encoding.
func EncodeLastPaywallAddressIndex(lp LastPaywallAddressIndex) ([]byte, error) {
	// make sure it user has record type and version specified
	lp.RecordType = RecordTypeLastPaywallAddrIdx
	lp.RecordVersion = DatabaseVersion

	b, err := json.Marshal(lp)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// DecodeLastPaywallAddressIndex decodes a JSON byte slice into a
// LastPaywallAddressIndex. It also adds the record type and version
// before encoding.
func DecodeLastPaywallAddressIndex(payload []byte) (*LastPaywallAddressIndex, error) {
	var lp LastPaywallAddressIndex

	err := json.Unmarshal(payload, &lp)
	if err != nil {
		return nil, err
	}

	err = verifyRecordVersion(lp.RecordVersion, DatabaseVersion)
	if err != nil {
		return nil, err
	}

	err = verifyRecordType(lp.RecordType, RecordTypeLastPaywallAddrIdx)
	if err != nil {
		return nil, err
	}

	return &lp, nil
}

// DecodeUser decodes a JSON byte slice into a User.
func DecodeUser(payload []byte) (*User, error) {
	var u User

	err := json.Unmarshal(payload, &u)
	if err != nil {
		return nil, err
	}

	err = verifyRecordVersion(u.RecordVersion, DatabaseVersion)
	if err != nil {
		return nil, err
	}

	err = verifyRecordType(u.RecordType, RecordTypeUser)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// EncodeEncryptionKey encodes EncryptionKey into a JSON byte slice.
func EncodeEncryptionKey(ek EncryptionKey) ([]byte, error) {
	k, err := json.Marshal(ek)
	if err != nil {
		return nil, err
	}

	return k, nil
}

// DecodeEncryptionKey decodes a JSON byte slice into EncryptionKey
func DecodeEncryptionKey(payload []byte) (*EncryptionKey, error) {
	var ek EncryptionKey

	err := json.Unmarshal(payload, &ek)
	if err != nil {
		return nil, err
	}

	return &ek, nil
}

// SaveEncryptionKey saves a EncryptionKey into the provided filename
func SaveEncryptionKey(ek EncryptionKey, filename string) error {
	k, err := EncodeEncryptionKey(ek)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, k, 0600)
}

// LoadEncryptionKey loads a EncryptionKey from the provided filename
func LoadEncryptionKey(filename string) (*EncryptionKey, error) {
	k, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ek, err := DecodeEncryptionKey(k)
	if err != nil {
		return nil, err
	}

	return ek, nil
}

// ResolveEncryptionKey creates and save a new encryption key in case
// there isn't one yet in the default home directory
func ResolveEncryptionKey(keyPath string) error {

	encryptionKeyPath := filepath.Join(keyPath, DefaultEncryptionKeyFilename)

	if !util.FileExists(encryptionKeyPath) {
		// create a new encryption key
		secretKey, err := sbox.NewKey()
		if err != nil {
			return err
		}

		err = SaveEncryptionKey(EncryptionKey{
			Key:  *secretKey,
			Time: time.Now().Unix(),
		}, encryptionKeyPath)
		if err != nil {
			return err
		}
	}
	return nil
}

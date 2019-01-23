// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package database

import (
	"errors"

	"github.com/decred/politeia/politeiad/api/v1/identity"
	"github.com/google/uuid"
)

type RecordTypeT int

var (
	// ErrUserNotFound indicates that a provided key was not found
	// in the database
	ErrNotFound = errors.New("key not found")

	// ErrUserExists indicates that a user already exists in the database.
	ErrUserExists = errors.New("user already exists")

	// ErrInvalidEmail indicates that a user's email is not properly formatted.
	ErrInvalidEmail = errors.New("invalid user email")

	// ErrShutdown is emitted when the database is shutting down.
	ErrShutdown = errors.New("database is shutting down")

	// ErrWrongVersion is emitted when the version in the database
	// does not match version of the interface implementation.
	ErrWrongVersion = errors.New("wrong database version")

	// ErrLoadingEncryptionKey is emitted when the encryption key cannot be
	// loaded from theprivded path
	ErrLoadingEncryptionKey = errors.New("encryption could not be loaded")
)

const (
	// DatabaseVersion is the current version of the database
	DatabaseVersion uint32 = 1

	// DatabaseVersionKey is the key used to map the database version
	DatabaseVersionKey = "userversion"

	// DefaultEncryptionKeyFilename is the name of the file where
	// the encryption key is stored
	DefaultEncryptionKeyFilename = "dbencryptionkey.json"

	LastPaywallAddressIndex = "lastpaywallindex"

	RecordTypeInvalid RecordTypeT = 0
	RecordTypeUser    RecordTypeT = 1
	RecordTypeVersion RecordTypeT = 2
)

// EncryptionKey wraps a key used for encrypting/decrypting the database
// data and the time when it was created
type EncryptionKey struct {
	Key  [32]byte // Key used for encryption
	Time int64    // Time key was created
}

// Identity wraps an ed25519 public key and timestamps to indicate if it is
// active.  If deactivated != 0 then the key is no longer valid.
type Identity struct {
	Key         [identity.PublicKeySize]byte // ed25519 public key
	Activated   int64                        // Time key was activated for use
	Deactivated int64                        // Time key was deactivated
}

// Version contains the database version.
type Version struct {
	RecordType    RecordTypeT `json:"recordtype"`
	RecordVersion uint32      `json:"recordversion"`

	Version uint32 `json:"version"` // Database version
	Time    int64  `json:"time"`    // Time of record creation
}

// A proposal paywall allows the user to purchase proposal credits.  Proposal
// paywalls are only valid for one tx.  The number of proposal credits created
// is determined by dividing the tx amount by the credit price.  Proposal
// paywalls expire after a set duration. politeiawww polls the paywall address
// for a payment tx until the paywall is either paid or it expires.
type ProposalPaywall struct {
	ID          uint64 // Paywall ID
	CreditPrice uint64 // Cost per proposal credit in atoms
	Address     string // Paywall address
	TxNotBefore int64  // Minimum timestamp for paywall tx
	PollExpiry  int64  // After this time, the paywall address will not be continuously polled
	TxID        string // Payment transaction ID
	TxAmount    uint64 // Amount sent to paywall address in atoms
	NumCredits  uint64 // Number of proposal credits created by payment tx
}

// A proposal credit allows the user to submit a new proposal.  Credits are
// created when a user sends a payment to a proposal paywall.  A credit is
// automatically spent when a user submits a new proposal.  When a credit is
// spent, it is updated with the proposal's censorship token and moved to the
// user's spent proposal credits list.
type ProposalCredit struct {
	PaywallID       uint64 // ID of the proposal paywall that created this credit
	Price           uint64 // Price this credit was purchased at in atoms
	DatePurchased   int64  // Unix timestamp of when the credit was purchased
	TxID            string // Payment transaction ID
	CensorshipToken string // Censorship token of proposal that used this credit
}

// User record.
type User struct {
	RecordType    RecordTypeT
	RecordVersion uint32

	ID                              uuid.UUID // Unique user uuid
	Email                           string    // Email address + lookup key.
	Username                        string    // Unique username
	HashedPassword                  []byte    // Blowfish hash
	Admin                           bool      // Is user an admin
	PaywallAddressIndex             uint64    // Sequential id used to generate paywall address
	NewUserPaywallAddress           string    // Address the user needs to send to
	NewUserPaywallAmount            uint64    // Amount the user needs to send
	NewUserPaywallTx                string    // Paywall transaction id
	NewUserPaywallTxNotBefore       int64     // Transactions occurring before this time will not be valid.
	NewUserPaywallPollExpiry        int64     // After this time, the user's paywall address will not be continuously polled
	NewUserVerificationToken        []byte    // New user registration verification token
	NewUserVerificationExpiry       int64     // New user registration verification expiration
	ResendNewUserVerificationExpiry int64     // Resend request for new user registration verification expiration
	UpdateKeyVerificationToken      []byte    // Verification token for updating keypair
	UpdateKeyVerificationExpiry     int64     // Verification expiration
	ResetPasswordVerificationToken  []byte    // Reset password token
	ResetPasswordVerificationExpiry int64     // Reset password token expiration
	LastLoginTime                   int64     // Unix timestamp of when the user last logged in
	FailedLoginAttempts             uint64    // Number of failed login a user has made in a row
	Deactivated                     bool      // Whether the account is deactivated or not
	EmailNotifications              uint64    // Notify the user via emails

	// Access times for proposal comments that have been accessed by the user.
	// Each string represents a proposal token, and the int64 represents the
	// time that the proposal has been most recently accessed in the format of
	// a UNIX timestamp.
	ProposalCommentsAccessTimes map[string]int64

	// All identities the user has ever used.  User should only have one
	// active key at a time.  We allow multiples in order to deal with key
	// loss.
	Identities []Identity

	// All proposal paywalls that have been issued to the user in chronological
	// order.
	ProposalPaywalls []ProposalPaywall

	// All proposal credits that have been purchased by the user, but have not
	// yet been used to submit a proposal.  Once a credit is used to submit a
	// proposal, it is updated with the proposal's censorship token and moved to
	// the user's spent proposal credits list.  The price that the proposal
	// credit was purchased at is in atoms.
	UnspentProposalCredits []ProposalCredit

	// All credits that have been purchased by the user and have already been
	// used to submit proposals.  Spent credits have a proposal censorship token
	// associated with them to signify that they have been spent. The price that
	// the proposal credit was purchased at is in atoms.
	SpentProposalCredits []ProposalCredit
}

// XXX Needs to be removed
// Database interface that is required by the web server.
// type Database interface {
// 	// User functions
// 	UserGet(string) (*User, error)           // Return user record, key is email
// 	UserGetByUsername(string) (*User, error) // Return user record given the username
// 	UserGetById(uuid.UUID) (*User, error)    // Return user record given its id
// 	UserNew(User) error                      // Add new user
// 	UserUpdate(User) error                   // Update existing user
// 	AllUsers(callbackFn func(u *User)) error // Iterate all users

// 	// Close performs cleanup of the backend.
// 	Close() error
// }

// Database interface
type Database interface {
	Put(string, []byte) error   // Set a value by key
	Get(string) ([]byte, error) // Get a database value by key

	Open() error
	Close() error
}

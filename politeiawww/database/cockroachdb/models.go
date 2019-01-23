package cockroachdb

// Version is the version of the cache the database is using.
type Version struct {
	Key       uint   `gorm:"primary_key"` // Primary key
	Version   uint32 `gorm:"not null"`    // Cache version
	Timestamp int64  `gorm:"not null"`    // UNIX timestamp of record creation
}

// User describes a key-value model for storing the user data
type User struct {
	UserID  string `gorm:"primary_key"` // Primary key
	Payload string `gorm:"not null"`    // String encoded user payload
}

func (Version) TableName() string {
	return tableVersion
}

func (User) TableName() string {
	return tableUsers
}

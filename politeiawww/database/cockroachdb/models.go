package cockroachdb

// KeyValue describes a key-value model for storing an encoded payload by key
type KeyValue struct {
	Key     string `gorm:"primary_key"` // Primary key
	Payload []byte `gorm:"not null"`    // Byte slice encoded payload
}

// TableName returns the table name for the KeyValue model
func (KeyValue) TableName() string {
	return tableKeyValue
}

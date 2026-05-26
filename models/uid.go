package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

const uidLength = 10
const uidCharset = "abcdefghijklmnopqrstuvwxyz"

// UID is a custom 10-digit lowercase English letter identifier used as the primary key for all entities.
type UID string

// NewUID generates a cryptographically random UID.
func NewUID() (UID, error) {
	letters := make([]byte, uidLength)
	for i := range letters {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(uidCharset))))
		if err != nil {
			return "", err
		}
		letters[i] = uidCharset[n.Int64()]
	}
	return UID(letters), nil
}

// String returns the string representation of a UID.
func (u UID) String() string {
	return string(u)
}

// Scan implements the sql.Scanner interface for UID.
func (u *UID) Scan(value any) error {
	switch v := value.(type) {
	case string:
		*u = UID(v)
	case []byte:
		*u = UID(v)
	case UID:
		*u = v
	default:
		return fmt.Errorf("cannot scan type %T into UID", value)
	}
	return nil
}

type Base struct {
	ID        UID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate generates a UID. This sets ID even if ID already has value. Because client might send IDs for creating
// This breaks seeding. You can do gorm SkipHooks if you want to seed. Make sure to set ID, CreatedAt etc because they'll be skipped
func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		id, err := NewUID()
		if err != nil {
			return err
		}
		b.ID = id
	}

	return nil
}

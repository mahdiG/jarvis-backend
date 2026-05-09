package models

import (
	"crypto/rand"
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
	*u = UID(value.(string))
	return nil
}

type Base struct {
	ID        UID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

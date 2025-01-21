package entity

import (
	"bytes"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Business struct {
	ID        string    `json:"id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Email    string `json:"email" gorm:"uniqueIndex;not null;"`
	Password []byte `json:"-"`
	Name     string `json:"name"`

	Promos []Promo `json:"promos" gorm:"foreignKey:CompanyID;"`
}

// HashedPassword is a function to hash the password.
func HashedPassword(password string) []byte {
	hashedPassword := argon2.IDKey([]byte(password), []byte("salt"), 1, 47104, 4, 32)
	return hashedPassword
}

// SetPassword is a method to hash the password before storing it.
func (business *Business) SetPassword(password string) {
	business.Password = HashedPassword(password)
}

// ComparePassword is a method to compare the password with the hashed password.
func (business *Business) ComparePassword(password string) error {
	if bytes.Equal(business.Password, HashedPassword(password)) {
		return nil
	} else {
		// to lazy to create a new error
		return bcrypt.ErrMismatchedHashAndPassword
	}
}

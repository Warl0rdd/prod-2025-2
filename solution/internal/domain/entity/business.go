package entity

import (
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
}

// HashedPassword is a function to hash the password.
func HashedPassword(password string) []byte {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return hashedPassword
}

// SetPassword is a method to hash the password before storing it.
func (business *Business) SetPassword(password string) {
	business.Password = HashedPassword(password)
}

// ComparePassword is a method to compare the password with the hashed password.
func (business *Business) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword(business.Password, []byte(password))
}

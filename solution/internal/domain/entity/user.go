package entity

import (
	"github.com/biter777/countries"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        string    `json:"id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Email     string                `json:"email" gorm:"index"`
	Password  []byte                `json:"-"`
	Name      string                `json:"name"`
	Surname   string                `json:"surname"`
	AvatarURL string                `json:"avatar_url"`
	Age       int                   `json:"age"`
	Country   countries.CountryCode `json:"country"`
}

// SetPassword is a method to hash the password before storing it.
func (user *User) SetPassword(password string) {
	user.Password = HashedPassword(password)
}

// ComparePassword is a method to compare the password with the hashed password.
func (user *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword(user.Password, []byte(password))
}

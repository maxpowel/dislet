package usermngr

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/pbkdf2"
	"crypto/sha1"
	"crypto/rand"
	"encoding/base64"
	"bytes"
	"fmt"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);not null;unique"`
	Password string `gorm:"not null"`
	Salt string `gorm:"not null"`
	Email string `gorm:"type:varchar(50);not null;unique"`
}


func NewUser() User {
	// Just initialize a random salt
	saltSize := 16
	salt := make([]byte, saltSize)
	rand.Read(salt)
	return User{
		Salt:base64.URLEncoding.EncodeToString(salt),
	}
}

func PlainPassword(user *User, plainPassword string) (error){
	// Salt must exists. If you created the user with NewUser there will not be any problem
	salt, err := base64.URLEncoding.DecodeString(user.Salt)
	if err != nil {
		return err
	}

	dk := pbkdf2.Key([]byte(plainPassword), salt, 4096, 32, sha1.New)
	user.Password = base64.URLEncoding.EncodeToString(dk)
	return nil
}

func CheckPassword(user *User, plainPassword string) (error) {
	salt, err := base64.URLEncoding.DecodeString(user.Salt)
	if err != nil {
		return err
	}

	password, err := base64.URLEncoding.DecodeString(user.Password)
	if err != nil {
		return err
	}
	dk := pbkdf2.Key([]byte(plainPassword), salt, 4096, 32, sha1.New)
	if bytes.Compare(dk, password) == 0 {
		return nil
	} else {
		return fmt.Errorf("Password does not match")
	}
}
package password

import (
	"golang.org/x/crypto/bcrypt"
)

type BcryptManager struct {
	salt string
}

// Для каждого пользователя должна быть своя соль, но пока так.
// Если останется время, то переделаю
func NewPasswordManager(salt string) *BcryptManager {
	return &BcryptManager{salt: salt}
}

func (h *BcryptManager) HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password+h.salt), bcrypt.DefaultCost)

	return string(hash)
}

func (h *BcryptManager) ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+h.salt))

	return err == nil
}

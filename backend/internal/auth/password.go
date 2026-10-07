package auth

import "golang.org/x/crypto/bcrypt"

const passwordCost = 12

type passwords struct{ dummyHash []byte }

func newPasswords() (passwords, error) {
	hash, err := hashPassword("dummy-password-for-timing")

	return passwords{dummyHash: []byte(hash)}, err
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)

	return string(hash), err
}

func (p passwords) check(hash, password string) bool {
	value := []byte(hash)
	if hash == "" {
		value = p.dummyHash
	}

	err := bcrypt.CompareHashAndPassword(value, []byte(password))

	return err == nil && hash != ""
}

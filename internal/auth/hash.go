package auth

import "golang.org/x/crypto/bcrypt"

func HashSecret(plainText string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainText), 12)
	return string(bytes), err
}

func CompareSecret(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

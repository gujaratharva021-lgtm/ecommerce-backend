package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// GenerateNumericOTP returns a random numeric code with the given number
// of digits (e.g. GenerateNumericOTP(6) -> "042917"), using
// crypto/rand so it's safe to use as a delivery-completion credential.
func GenerateNumericOTP(digits int) (string, error) {
	if digits <= 0 {
		digits = 6
	}
	max := int64(1)
	for i := 0; i < digits; i++ {
		max *= 10
	}
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", digits, n.Int64()), nil
}

// HashOTP hashes a plaintext OTP for storage. Only this hash is ever
// persisted - the plaintext code must never be written to the database.
func HashOTP(code string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CompareOTP reports whether the plaintext code matches a previously
// hashed OTP.
func CompareOTP(hash, code string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil
}

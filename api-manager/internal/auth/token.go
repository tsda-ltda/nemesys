package auth

import (
	"crypto/rand"
	"math/big"
)

// Chars used to generate a token.
const tokenChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-!$%#@+"

// NewToken returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func NewToken(n int) (string, error) {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenChars))))
		if err != nil {
			return "", err
		}
		ret[i] = tokenChars[num.Int64()]
	}

	return string(ret), nil
}

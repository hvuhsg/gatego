package security

import (
	"crypto/rand"
	"math/big"
)

var charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateRandomString(length int) string {
	randomString := make([]byte, length)
	max := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err) // Handle error appropriately
		}
		randomString[i] = charset[randomIndex.Int64()]
	}
	return string(randomString)
}

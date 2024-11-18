package security

import (
	"crypto/sha256"
	"encoding/hex"
)

const salt = "somesaltvalueamiright"

func HashWithSalt(value string) string {
	// Concatenate the password with the salt
	saltedValue := value + salt

	// Create a new SHA256 hash object
	hasher := sha256.New()

	// Write the salted password bytes to the hash object
	hasher.Write([]byte(saltedValue))

	// Get the hashed password bytes
	hashedPassword := hasher.Sum(nil)

	// Convert the hashed password bytes to a hexadecimal string
	hashedPasswordString := hex.EncodeToString(hashedPassword)

	return hashedPasswordString
}

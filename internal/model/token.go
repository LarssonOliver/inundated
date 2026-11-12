package model

import "crypto/sha256"

// HashToken is how session tokens and app passwords are stored: only the hash
// is kept, so a leaked database does not leak credentials.
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

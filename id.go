package thing2

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomId generates a hex-encoded 8-byte random ID in format
// 'xxxxxxxx-xxxxxxxx'
func GenerateRandomId() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	hexString := hex.EncodeToString(bytes)
	return hexString[:8] + "-" + hexString[8:]
}

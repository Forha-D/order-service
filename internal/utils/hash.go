package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func SHA256(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func BuildIdempotencyKey(userID, operation, clientKey string) string {
	canonical := strings.TrimSpace(userID) + ":" + strings.TrimSpace(operation) + ":" + strings.TrimSpace(clientKey)
	return SHA256(canonical)
}

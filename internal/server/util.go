package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

func convertLabelsToTags(labels map[string]string) string {
	const sep = ","

	var tags []string
	for key, value := range labels {
		if !strings.Contains(key, sep) && !strings.Contains(value, sep) {
			tags = append(tags, fmt.Sprintf("%s = %s", key, value))
		}
	}

	return strings.Join(tags, sep)
}

// generateRequestID generates a random 32 character long request ID for use
// with log line correlation. If reading from the system CSPRNG fails, "nil" is
// returned.
func generateRequestID() string {
	const len = 16

	bytes := make([]byte, len)
	if _, err := rand.Read(bytes); err != nil {
		return "nil"
	}

	return hex.EncodeToString(bytes)
}

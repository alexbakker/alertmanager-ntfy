package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/alexbakker/alertmanager-ntfy/internal/alertmanager"
)

const (
	tagSeparator = ","
)

// templateContext is the data passed to notification templates. It contains
// the individual alert as well as the parent webhook payload.
type templateContext struct {
	*alertmanager.Alert
	Payload *alertmanager.Payload
}

// exprMap builds a map for gval expression evaluation. Alert fields are
// available at the top level for backwards compatibility, as well as under the
// "alert" key. The parent webhook payload is available under "payload".
func exprMap(alert *alertmanager.Alert, payload *alertmanager.Payload) map[string]interface{} {
	m := alert.Map()
	m["alert"] = alert.Map()
	m["payload"] = payload.Map()
	return m
}

func convertLabelsToTags(labels map[string]string) []string {
	var tags []string
	for key, value := range labels {
		if !strings.Contains(key, tagSeparator) && !strings.Contains(value, tagSeparator) {
			tags = append(tags, fmt.Sprintf("%s = %s", key, value))
		}
	}

	return tags
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

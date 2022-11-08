package main

import (
	"fmt"
	"os"
)

// getDefaultTopic generates a default topic starting with alertmanager-ntfy-
// and the system's hostname appended after it.
func getDefaultTopic() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "error"
	}

	return fmt.Sprintf("alertmanager-ntfy-%s", hostname)
}

package api

import "fmt"

// logVerbose writes to VerboseLog if it's set
func (c *Client) logVerbose(format string, args ...interface{}) {
	if c.VerboseLog != nil {
		fmt.Fprintf(c.VerboseLog, format, args...)
	}
}

// truncateStringUTF8Safe safely truncates a string to maxRunes runes,
// ensuring we don't split multi-byte UTF-8 characters.
func truncateStringUTF8Safe(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

package util

import "strings"

// NormalizeAndValidatePlayerPosition checks if the provided position is one of the allowed values (case-insensitively)
// and returns the canonical version of the position if valid.
func NormalizeAndValidatePlayerPosition(position string) (string, bool) {
	// Using a map for efficient lookup. Keys are lowercase for case-insensitive matching.
	// Values are the canonical representation.
	allowedPositions := map[string]string{
		"penyerang":      "Penyerang",
		"gelandang":      "Gelandang",
		"bertahan":       "Bertahan",
		"penjaga gawang": "Penjaga Gawang",
	}

	canonicalPosition, ok := allowedPositions[strings.ToLower(position)]
	return canonicalPosition, ok
}

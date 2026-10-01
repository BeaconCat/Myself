package updater

import (
	"regexp"
	"strings"
)

// Display names and prerelease labels do not participate in numeric ordering.
// Decimal strings avoid overflow for version components larger than uint64.
var versionRE = regexp.MustCompile(`^[vV]?[0-9]+(?:\.[0-9]+){1,15}(?:-[0-9A-Za-z][0-9A-Za-z.-]*)?$`)

func NumericVersion(version string) (string, bool) {
	version = strings.TrimSpace(version)
	if len(version) > 256 || !versionRE.MatchString(version) {
		return "", false
	}
	version = strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V")
	version, _, _ = strings.Cut(version, "-")
	parts := strings.Split(version, ".")
	for i, part := range parts {
		parts[i] = strings.TrimLeft(part, "0")
		if parts[i] == "" {
			parts[i] = "0"
		}
	}
	for len(parts) > 1 && parts[len(parts)-1] == "0" {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, "."), true
}

func CompareVersions(a, b string) (int, bool) {
	a, okA := NumericVersion(a)
	b, okB := NumericVersion(b)
	if !okA || !okB {
		return 0, false
	}
	x, y := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(x), len(y)); i++ {
		left, right := "0", "0"
		if i < len(x) {
			left = x[i]
		}
		if i < len(y) {
			right = y[i]
		}
		if len(left) < len(right) {
			return -1, true
		}
		if len(left) > len(right) {
			return 1, true
		}
		if left < right {
			return -1, true
		}
		if left > right {
			return 1, true
		}
	}
	return 0, true
}

func newer(a, b string) bool { n, ok := CompareVersions(a, b); return ok && n > 0 }

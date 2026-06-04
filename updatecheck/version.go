// Package updatecheck reports whether a newer GitHub release of a CLI exists and
// how to upgrade it via the package manager that installed it. It never touches
// the binary — the actual upgrade is delegated to the package manager (see
// ADR-0005).
package updatecheck

import (
	"strconv"
	"strings"
)

// isNewer reports whether latest is strictly greater than current. Versions are
// compared component-by-component as dot-separated integers after stripping an
// optional leading "v" — this works for both SemVer (1.2.3) and CalVer
// (2026.5.0). An empty latest returns false (never suggest an update).
func isNewer(latest, current string) bool {
	l := strings.TrimPrefix(latest, "v")
	if l == "" {
		return false
	}
	return compareVersions(l, strings.TrimPrefix(current, "v")) > 0
}

// compareVersions returns -1, 0, or 1 comparing dot-separated integer versions.
// Missing trailing components are treated as 0 (1.2 == 1.2.0); non-numeric
// components compare as 0.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := range max(len(aParts), len(bParts)) {
		var av, bv int
		if i < len(aParts) {
			av, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bv, _ = strconv.Atoi(bParts[i])
		}
		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		}
	}
	return 0
}

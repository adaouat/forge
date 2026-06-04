package updatecheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"v1.3.0", "v1.2.3", true},
		{"1.3.0", "1.2.3", true},
		{"v1.2.3", "v1.2.3", false},     // equal
		{"v1.2.0", "v1.3.0", false},     // older
		{"v1.10.0", "v1.9.0", true},     // numeric, not lexical
		{"v2.0.0", "1.9.9", true},       // mixed prefix
		{"2026.5.0", "2026.4.9", true},  // CalVer
		{"2026.4.0", "2026.5.0", false}, // CalVer older
		{"", "v1.0.0", false},           // empty latest never suggests an update
		{"v1.0.0", "dev", true},         // non-version current → any release is newer
	}
	for _, tc := range tests {
		t.Run(tc.latest+"_vs_"+tc.current, func(t *testing.T) {
			assert.Equal(t, tc.want, isNewer(tc.latest, tc.current))
		})
	}
}

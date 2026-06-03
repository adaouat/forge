package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adaouat/forge/ui"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want ui.Mode
	}{
		{"human", "human", ui.Human},
		{"plain", "plain", ui.Plain},
		{"json", "json", ui.JSON},
		{"empty defaults to human", "", ui.Human},
		{"unknown defaults to human", "bogus", ui.Human},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ui.ParseMode(tc.in))
		})
	}
}

func TestMode_IsHuman(t *testing.T) {
	assert.True(t, ui.Human.IsHuman())
	assert.False(t, ui.Plain.IsHuman())
	assert.False(t, ui.JSON.IsHuman())
}

func TestMode_ZeroValueIsHuman(t *testing.T) {
	var m ui.Mode
	assert.True(t, m.IsHuman(), "zero value must be human (bifrost's historic default)")
}

func TestMode_String(t *testing.T) {
	assert.Equal(t, "human", ui.Human.String())
	assert.Equal(t, "plain", ui.Plain.String())
	assert.Equal(t, "json", ui.JSON.String())
}

func TestMode_StringRoundTripsParseMode(t *testing.T) {
	for _, m := range []ui.Mode{ui.Human, ui.Plain, ui.JSON} {
		assert.Equal(t, m, ui.ParseMode(m.String()))
	}
}

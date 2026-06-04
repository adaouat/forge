package updatecheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name, path, goBin string
		want              InstallMethod
	}{
		{"homebrew arm", "/opt/homebrew/Cellar/heraut/0.4.0/bin/heraut", "", Homebrew},
		{"homebrew intel", "/usr/local/Cellar/heraut/0.4.0/bin/heraut", "", Homebrew},
		{"linuxbrew", "/home/linuxbrew/.linuxbrew/Cellar/heraut/0.4.0/bin/heraut", "", Homebrew},
		{"mise github backend", "/home/u/.local/share/mise/installs/github-adaouat-heraut/0.4.0/heraut", "", Mise},
		{"scoop windows", `C:\Users\u\scoop\apps\heraut\current\heraut.exe`, "", Scoop},
		{"go install gobin", "/home/u/go/bin/heraut", "/home/u/go/bin", GoInstall},
		{"manual in usr local bin", "/usr/local/bin/heraut", "", Unknown},
		{"manual in home local bin", "/home/u/.local/bin/heraut", "", Unknown},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, detect(tc.path, tc.goBin))
		})
	}
}

func TestUpgradeCommand(t *testing.T) {
	assert.Equal(t, "brew upgrade heraut", Homebrew.UpgradeCommand("heraut", ""))
	assert.Equal(t, "mise upgrade heraut", Mise.UpgradeCommand("heraut", ""))
	assert.Equal(t, "scoop update heraut", Scoop.UpgradeCommand("heraut", ""))
	assert.Equal(t, "go install github.com/adaouat/heraut/cmd/heraut@latest",
		GoInstall.UpgradeCommand("heraut", "github.com/adaouat/heraut/cmd/heraut"))
	assert.Equal(t, "", Unknown.UpgradeCommand("heraut", ""), "unknown has no command; caller falls back")
}

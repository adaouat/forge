package updatecheck

import (
	"os"
	"path/filepath"
	"strings"
)

// InstallMethod is how a binary was installed, inferred from its path.
type InstallMethod int

const (
	Unknown InstallMethod = iota
	Homebrew
	Mise
	Scoop
	GoInstall
)

// DetectInstall infers how the running executable was installed by resolving its
// real path and matching known package-manager locations.
func DetectInstall() InstallMethod {
	exe, err := os.Executable()
	if err != nil {
		return Unknown
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return detect(exe, goBinDir())
}

// detect classifies an executable path. goBin is the resolved `go install`
// destination ("" if unknown), used to recognise go-installed binaries.
func detect(exePath, goBin string) InstallMethod {
	switch {
	case strings.Contains(exePath, "/Cellar/"),
		strings.Contains(exePath, "/homebrew/"),
		strings.Contains(exePath, "/linuxbrew/"):
		return Homebrew
	case strings.Contains(exePath, "/mise/installs/"):
		return Mise
	case strings.Contains(exePath, `\scoop\`),
		strings.Contains(exePath, "/scoop/"):
		return Scoop
	case goBin != "" && filepath.Dir(exePath) == goBin:
		return GoInstall
	}
	return Unknown
}

func goBinDir() string {
	if b := os.Getenv("GOBIN"); b != "" {
		return b
	}
	if p := os.Getenv("GOPATH"); p != "" {
		return filepath.Join(p, "bin")
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, "go", "bin")
	}
	return ""
}

// UpgradeCommand returns the command to upgrade bin via this install method, or
// "" for Unknown (callers fall back to a generic message). module is the
// `go install` target path, used only for GoInstall.
func (m InstallMethod) UpgradeCommand(bin, module string) string {
	switch m {
	case Homebrew:
		return "brew upgrade " + bin
	case Mise:
		return "mise upgrade " + bin
	case Scoop:
		return "scoop update " + bin
	case GoInstall:
		return "go install " + module + "@latest"
	default:
		return ""
	}
}

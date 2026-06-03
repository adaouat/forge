package ui

// HelpLong returns ASCII art followed by a catch-phrase, for use as cobra
// root.Long. Apps pass their own art and tagline.
func HelpLong(art, catchPhrase string) string {
	return art + "\n" + catchPhrase
}

// VersionTemplate returns a cobra text/template string for --version output.
// cobra fills {{.Name}} and {{.Version}} at runtime.
func VersionTemplate(art, catchPhrase string) string {
	return art + "\n\n  " + catchPhrase + "\n\n  {{.Name}} {{.Version}}\n\n"
}

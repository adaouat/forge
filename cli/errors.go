package cli

import (
	"io"
	"strings"
	"unicode"

	"charm.land/fang/v2"
)

// errorHandler renders errors with fang's default layout but without its title-casing of the
// first word, which rewrites a leading flag or tag (--pre-release → --Pre-Release).
func errorHandler(w io.Writer, styles fang.Styles, err error) {
	styles.ErrorText = styles.ErrorText.Transform(capitalizeLeadingWord)
	fang.DefaultErrorHandler(w, styles, err)
}

// capitalizeLeadingWord upper-cases the first letter of s when its first word is plain letters
// (optionally ending in punctuation such as a colon); any other first word is an identifier and
// is returned untouched.
func capitalizeLeadingWord(s string) string {
	runes := []rune(s)
	start := 0
	for start < len(runes) && unicode.IsSpace(runes[start]) {
		start++
	}
	end := start
	for end < len(runes) && !unicode.IsSpace(runes[end]) {
		end++
	}
	word := strings.TrimRight(string(runes[start:end]), ":,;.")
	if word == "" {
		return s
	}
	for _, r := range word {
		if !unicode.IsLetter(r) {
			return s
		}
	}
	runes[start] = unicode.ToUpper(runes[start])
	return string(runes)
}

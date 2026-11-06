package cli

import (
	"io"
	"strings"
	"unicode"

	"charm.land/fang/v2"
)

// errorHandler renders errors with fang's default layout but without its title-casing of the
// first word, which rewrites a leading flag or tag (--pre-release → --Pre-Release), and without
// the second period it adds to a message that already ends with punctuation.
func errorHandler(w io.Writer, styles fang.Styles, err error) {
	styles.ErrorText = styles.ErrorText.Transform(func(s string) string {
		return dropRedundantPeriod(capitalizeLeadingWord(s))
	})
	fang.DefaultErrorHandler(w, styles, err)
}

// dropRedundantPeriod removes the final period when it follows ".", "!", "?" or "…". fang renders
// err.Error()+"." and runs the transform on that whole string, so the period it appended is the
// last rune, after the final line of a multi-line message.
func dropRedundantPeriod(s string) string {
	runes := []rune(s)
	n := len(runes)
	if n < 2 || runes[n-1] != '.' {
		return s
	}
	switch runes[n-2] {
	case '.', '!', '?', '…':
		return string(runes[:n-1])
	}
	return s
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

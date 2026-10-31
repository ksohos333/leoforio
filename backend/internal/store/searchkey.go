package store

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// SearchKey normalises a name so that it can be searched without regard to
// case or accents.
//
// Greek is the case that matters here: the accented vowels carry a mark
// that a person may or may not type, and the final sigma has two forms that
// mean the same thing at the end of a word.
func SearchKey(value string) string {
	decomposed := norm.NFD.String(value)

	var builder strings.Builder
	builder.Grow(len(decomposed))

	for _, letter := range decomposed {
		if unicode.Is(unicode.Mn, letter) {
			continue
		}
		lower := unicode.ToLower(letter)
		if lower == 'ς' {
			lower = 'σ'
		}
		builder.WriteRune(lower)
	}

	return builder.String()
}

// escapeLike escapes the wildcards of a LIKE pattern so that a query
// containing them searches for them literally.
func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}

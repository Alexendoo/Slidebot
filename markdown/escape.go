package markdown

import (
	"strings"
)

func newEscaper(chars ...string) *strings.Replacer {
	replacements := []string{}

	for _, char := range chars {
		replacements = append(replacements, char, "\\"+char)
	}

	return strings.NewReplacer(replacements...)
}

var escaper = newEscaper(
	"*", "_", "~", // basic formatting
	"`",                // code blocks
	"<", ">", "@", "#", // Discord mentions
	":", // Emoji
)

// Escape a string for inclusion within discord markdown
func Escape(s string) string {
	return escaper.Replace(s)
}

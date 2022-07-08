package core

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func toCamelCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

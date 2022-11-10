package grammar

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var immutableWords = map[string]bool{
	"sheep":  true,
	"deer":   true,
	"fish":   true,
	"moose":  true,
	"shrimp": true,
	"data":   true,
}

var irregularWords = map[string]string{
	"child":  "children",
	"man":    "men",
	"woman":  "women",
	"goose":  "geese",
	"person": "people",
	"tooth":  "teeth",
	"foot":   "feet",
	"mouse":  "mice",
}

var exceptions = map[string]string{
	"photo":  "photos",
	"roof":   "roofs",
	"belief": "beliefs",
	"chef":   "chefs",
	"chief":  "chiefs",
	"quiz":   "quizzes",
}

// Pluralize returns a plural (and lowered) form of a word. This is a naive approach based on
// https://www.lingobest.com/free-online-english-course/plural-nouns-in-english/ but aspire to be sufficient for
// names used in e.g. APIs.
func Pluralize(word string) string {

	toLower := cases.Lower(language.AmericanEnglish)
	w := toLower.String(word)

	if immutableWords[w] {
		return w
	}

	if res, ok := irregularWords[w]; ok {
		return res
	}

	if res, ok := exceptions[w]; ok {
		return res
	}

	l := len(w)
	lastTwo := w[l-2 : l]

	switch lastTwo {
	case "ch", "sh":
		return w + "es"
	case "fe":
		return w[:l-2] + "ves"
	case "us":
		// NOTE: This seems like not necessary for US English. (cactus - cacti, focus - foci, etc..)
		//return w[:l-2] + "i"
	}

	lastOne := w[l-1 : l]

	switch lastOne {
	case "s", "x", "z":
		return w + "es"
	case "y":
		if lastTwo != "oy" && lastTwo != "ey" && lastTwo != "ay" {
			return w[:l-1] + "ies"
		}
	case "f":
		return w[:l-1] + "ves"
	case "o":
		if lastTwo != "oo" && lastTwo != "io" {
			return w + "es"
		}
	}

	return w + "s"
}

package grammar

import (
	"testing"
)

func TestPluralize(t *testing.T) {

	feed := map[string]string{
		"sofa":     "sofas",
		"computer": "computers",
		"phone":    "phones",
		"car":      "cars",
		"book":     "books",
		"chair":    "chairs",
		"bus":      "buses",
		"box":      "boxes",
		"witch":    "witches",
		"quiz":     "quizzes",
		"brush":    "brushes",
		"dataset":  "datasets",
		"photo":    "photos",
		"baby":     "babies",
		"city":     "cities",
		"library":  "libraries",
		"lady":     "ladies",
		"berry":    "berries",
		"boy":      "boys",
		"play":     "plays",
		"ray":      "rays",
		"turkey":   "turkeys",
		"monkey":   "monkeys",
		"tray":     "trays",
		"wife":     "wives",
		"calf":     "calves",
		"elf":      "elves",
		"loaf":     "loaves",
		"self":     "selves",
		"roof":     "roofs",
		"zoo":      "zoos",
		"studio":   "studios",
		"potato":   "potatoes",
		"tomato":   "tomatoes",
		"mosquito": "mosquitoes",
		"domino":   "dominoes",
		"hero":     "heroes",
		"sheep":    "sheep",
		"child":    "children",
	}

	for k, v := range feed {
		t.Run(k, func(t *testing.T) {
			result := Pluralize(k)
			if result != v {
				t.Errorf("expected %q, got %q", v, result)
			}
		})
	}
}

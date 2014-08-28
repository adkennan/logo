package main

import (
	"testing"
)

func assertWords(t *testing.T, fn Node, err error, expected ...string) {

	if err != nil {
		t.Errorf("Error \"%s\"", err)
		return
	}

	e := EnumerateWords(fn)
	for ix, v := range expected {
		wn := e.nextWord()
		if v == "" {
			if wn != nil {
				t.Errorf("%d: Expected nil.", ix)
				return
			}
		} else {
			if wn == nil {
				t.Errorf("%d: No value.", ix)
				return
			}
			if wn.value != v {
				t.Errorf("%d:Expected \"%s\" was \"%s\"", ix, v, wn.value)
				return
			}
		}
	}
}

func TestEnumerateSingleWord(t *testing.T) {

	n, err := ParseString("Hello")
	assertWords(t, n, err, "Hello")
}

func TestEnumerateMultipleWord(t *testing.T) {

	n, err := ParseString("Goodbye Cruel World")
	assertWords(t, n, err, "Goodbye", "Cruel", "World")
}

func TestEnumerateList(t *testing.T) {
	n, err := ParseString("[ Hello World ]")

	assertWords(t, n, err, "Hello", "World")
}

func TestEnumerateEmbeddedList(t *testing.T) {
	n, err := ParseString("Say [ Hello World ] Again")

	assertWords(t, n, err, "Say", "Hello", "World", "Again")
}

func TestEnumerateEmbeddedLists(t *testing.T) {
	n, err := ParseString("Say [ Hello [ World ] ] Again")

	assertWords(t, n, err, "Say", "Hello", "World", "Again")
}

func TestEnumerateGroup(t *testing.T) {
	n, err := ParseString("Say ( Hello [ World  ] Again )")

	assertWords(t, n, err, "Say", "Hello", "World", "Again")
}

func TestEnumerateEmptyList(t *testing.T) {
	n, err := ParseString("Say [ Hello [ [ ]  ]  World ] Again")

	assertWords(t, n, err, "Say", "Hello", "World", "Again")
}

func TestEnumerateUntilNil(t *testing.T) {

	n, err := ParseString("Say [ Hello [ [ ]  ]  World ] Again")

	assertWords(t, n, err, "Say", "Hello", "World", "Again", "")
}

func TestEnumerateExpression(t *testing.T) {

	n, err := ParseString("2 * -5")

	assertWords(t, n, err, "2", "*", "-5")
}

func TestEnumerateMultipleLines(t *testing.T) {

	n, err := ParseString("  circle :n 1\n  circle :n -1\n  circs :n / 2")

	assertWords(t, n, err, "circle", ":n", "1", "circle", ":n", "-1", "circs", ":n", "/", "2")

}

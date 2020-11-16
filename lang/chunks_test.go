package lang

import (
	"testing"
)

func TestSplitter(t *testing.T) {
	s := "abc def"
	expected := []string{
		"a", "ab", "abc", "abc d", "abc de", "d", "de", "def",
	}
	chunks := Splitter(s, 0)
	if len(chunks) != len(expected) {
		t.Errorf("have lenght: %d; expected: %d", len(chunks), len(expected))
	}
	for k, chunk := range chunks {
		if expected[k] != chunk {
			t.Errorf("got: %s; expected: %s", chunk, expected[k])
		}
	}
}

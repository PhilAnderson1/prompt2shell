package main

import (
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"\n", true},
		{"yes\n", false},
		{" \n", false},
		{"", false},
		{"\x1b", false},
	}
	for _, test := range tests {
		got, err := confirm(strings.NewReader(test.input))
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("confirm(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}

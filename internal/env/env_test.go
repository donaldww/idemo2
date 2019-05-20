package env

import (
	"testing"
)

func TestHome(t *testing.T) {

	var simpleTests = []struct {
		in  string
		out string
	}{
		{Home(), "/usr/local/ig"},
		{Bin(), "/usr/local/ig/bin"},
		{Config(), "/usr/local/ig/config"},
		{Data(), "/usr/local/ig/data"},
		{Tmp(), "/usr/local/ig/tmp"},
	}

	for _, tt := range simpleTests {
		t.Run(tt.in, func(t *testing.T) {
			if tt.in != tt.out {
				t.Errorf("got %q, wanted %q", tt.in, tt.out)
			}
		})

	}
}

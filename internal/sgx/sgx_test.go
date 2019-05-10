// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package sgx

import (
	"fmt"
	"math"
	"os"
	"testing"
	
	"github.com/donaldww/ig"
)

func TestPrint(t *testing.T) {
	str := "Hello World"
	want := len(str)
	got, err := fmt.Print(str)
	if err != nil {
		t.FailNow()
	}
	if want != got {
		t.Errorf("Println: got %d, wanted %d", got, want)
	}
}

func TestAbs(t *testing.T) {
	got := math.Abs(-1.0)
	if got != 1 {
		t.Errorf("Abs(-1) = %f; want 1", got)
	}
}

func TestInfiniBin(t *testing.T) {
	home, _ := os.UserHomeDir()
	homeDir := home + "/" + ".infinigon/bin"

	var tests = []struct {
		want string
	}{
		{homeDir},
	}

	for _, test := range tests {
		got := ig.Env("IGBIN")
		if got != test.want {
			t.Errorf("infiniBin: %s != %s", got, test.want)
		}
	}
}

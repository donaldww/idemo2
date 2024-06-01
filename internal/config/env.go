// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package config

import (
	"log"
	"os"
)

type Home struct {
	home string `env:"HOME"`
}

var myHome Home

func init() {
	d, ok := os.LookupEnv("ENCLAVE_HOME")
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		myHome.home = home + "/.config/enclave"
	} else {
		myHome.home = d
	}
}

// HomeConfig returns the configuration directory.
func HomeConfig() string {
	return myHome.home + "/config"
}

// Bin returns IG bin directory.
func Bin() string {
	return myHome.home + "/bin"
}

// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package config

import (
	"log"
	"os"
)

var igHome string

func init() {
	d, ok := os.LookupEnv("IGHOME")
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		igHome = home + "/ig"
	} else {
		igHome = d
	}
}

// Tmp returns IG tmp dir
func Tmp() string {
	d, ok := os.LookupEnv("IGTMP")
	if !ok {
		return igHome + "/tmp"
	} else {
		return d
	}
}

// Data returns the data directory.
func Data() string {
	d, ok := os.LookupEnv("IGDATA")
	if !ok {
		return igHome + "/data"
	} else {
		return d
	}
}

// HomeConfig returns the configuration directory.
func HomeConfig() string {
	return igHome + "/config"
}

// Bin returns IG bin directory.
func Bin() string {
	return igHome + "/bin"
}

// Home returns IG home directory.
func Home() string {
	return igHome
}

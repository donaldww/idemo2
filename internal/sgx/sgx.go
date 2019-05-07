// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package sgx

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Get the state of the enclave when the program starts.
func init() {
	Scan()
}

func IsInfinibinExist() bool {
	ibin, err := infiniBin()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(ibin); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Scan scans the Infinigon SGX enclave binaries.
func Scan() {
	path, err := infiniBin()
	if err != nil {
		log.Panic(err)
	}

	err = filepath.Walk(path, walk)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// infiniBin returns a path to user's infinigon 'bin' directory.
func infiniBin() (ret_ string, err_ error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	s := []string{home, ".infinigon/bin"}
	ret_ = strings.Join(s, "/")
	err_ = nil
	return
}

// walk process information about each file and directory in the SGX enclave.
func walk(_path string, _info os.FileInfo, _err error) (err_ error) {
	_, _ = _info, _err
	err_ = nil

	fileInfo, err := os.Stat(_path)
	if err != nil {
		return
	}

	mode := fileInfo.Mode()
	if mode.IsRegular() {
		fmt.Println("f", _path)
		return
	}

	if mode.IsDir() {
		fmt.Println("d", _path)
		return
	}

	fmt.Println(_path)
	return
}

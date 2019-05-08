// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package sgx

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// enclaveItem represents a file or directory in the enclave.
type enclaveItem struct {
	Path   string
	Type   string
	Md5    string
	Shasum string
}

// println prints an enclave item.
func (e enclaveItem) println() {
	fmt.Println(e.Path, e.Type, e.Md5, e.Shasum)
}

var stableEnclave = make([]enclaveItem, 0)
var scannedEnclave = make([]enclaveItem, 0)
var currentEnclave *[]enclaveItem

// Get the state of the enclave when the program starts.
func init() {
	currentEnclave = &stableEnclave
	Scan()
	// PrintStable()
	currentEnclave = &scannedEnclave
}

// func PrintStable() {
// 	fmt.Println("STABLE ENCLAVE")
// 	for _, e := range stableEnclave {
// 		e.Println()
// 	}
// }

// Scan scans the Infinigon SGX enclave binaries.
func Scan() {
	path, err := InfiniBin()
	if err != nil {
		log.Panic(err)
	}
	err = filepath.Walk(path, walk)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Reset the scannedEnclave to nil before the next run.
func Reset() {
	scannedEnclave = nil
}

func PrintScanned() {
	fmt.Println("SCANNED ENCLAVE")
	for _, x := range scannedEnclave {
		x.println()
	}
	Reset()
}

// InfiniBin returns a path to user's infinigon 'bin' directory.
func InfiniBin() (ret_ string, err_ error) {
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
	shasum := fmt.Sprintf("%x", sha256.Sum256([]byte(_path)))

	fileInfo, err := os.Stat(_path)
	if err != nil {
		return
	}
	mode := fileInfo.Mode()

	if mode.IsRegular() {
		data, err2 := ioutil.ReadFile(_path)
		if err2 != nil {
			log.Panic(err2)
		}
		newMd5 := fmt.Sprintf("%x", md5.Sum(data))
		*currentEnclave = append(*currentEnclave, enclaveItem{_path, "f", newMd5, shasum})
	} else if mode.IsDir() {
		*currentEnclave = append(*currentEnclave, enclaveItem{_path, "d", "", shasum})
	} else {
		*currentEnclave = append(*currentEnclave, enclaveItem{_path, "o", "", shasum})
	}

	err_ = nil
	return
}

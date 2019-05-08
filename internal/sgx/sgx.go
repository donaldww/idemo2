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
	"time"
)

// enclaveItem represents a file or directory in the enclave.
type enclaveItem struct {
	Type   string
	Md5    string
	Shasum string
}

type enclaveMap map[string]enclaveItem

var stableEnclave = enclaveMap{}
var scannedEnclave = enclaveMap{}
var oneTime = true

// println prints an enclave item.
func (e enclaveItem) println() {
	fmt.Println(e.Type, e.Md5, e.Shasum)
}

// Get the state of the enclave when the program starts.
func init() {
	oneTime = true
	Scan()
	oneTime = false
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

// walk process information about each file and directory in the SGX enclave.
func walk(_path string, _info os.FileInfo, _err error) (err_ error) {
	_, _ = _info, _err
	shasum := fmt.Sprintf("%x", sha256.Sum256([]byte(_path)))

	fileInfo, err := os.Stat(_path)
	if err != nil {
		return
	}
	mode := fileInfo.Mode()

	switch {
	case mode.IsRegular():
		newMd5 := func(p string) string {
			data, err2 := ioutil.ReadFile(p)
			if err2 != nil {
				log.Panic(err2)
			}
			return fmt.Sprintf("%x", md5.Sum(data))
		}(_path)
		if oneTime {
			stableEnclave[_path] = enclaveItem{Type: "f", Md5: newMd5, Shasum: shasum}
		} else {
			scannedEnclave[_path] = enclaveItem{Type: "f", Md5: newMd5, Shasum: shasum}
		}
	case mode.IsDir():
		// break
		// stableEnclave[_path] = enclaveItem{Type: "d", Md5: "", Shasum: shasum}
	default:
		// break
		// stableEnclave[_path] = enclaveItem{Type: "u", Md5: "", Shasum: shasum}
	}

	err_ = nil
	return
}

/*************
	Errors
**************/

// EnclaveError is used to express enclave errors.
type enclaveError struct {
	When time.Time
	What string
}

func (e enclaveError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

// IsValid determines if a scanned directory matches a valid one.
func IsValid() (err_ error) {
	err_ = nil

	fmt.Println("Valid Enclave:", len(stableEnclave))
	fmt.Println("Scanned Enclave:", len(scannedEnclave))

	if len(stableEnclave) != len(scannedEnclave) {
		return enclaveError{
			time.Date(1989, 3, 15, 22, 30, 0, 0, time.UTC),
			"File mismatch: FAIL",
		}
	}
	return
}

// Reset the scannedEnclave to nil before the next run.
func Reset() {
	scannedEnclave = enclaveMap{}
}

// InfiniBin returns a path to user's infinigon 'bin' directory.
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

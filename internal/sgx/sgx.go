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
	"time"
	
	"github.com/donaldww/ig"
)

// enclaveItem represents a file or directory in the enclave.
type enclaveItem struct {
	Name   string
	Path   string
	Type   string
	Md5    string
	Shasum string
}

type enclaveMap map[string]enclaveItem

// This is the enclave that is scanned every time.
var scannedEnclave = enclaveMap{}
var scannedList []string = nil

// This is a copy of the first scanned enclave.
var stableEnclave = enclaveMap{}
var stableList []string = nil

// var oneTime = true

// println prints an enclave item.
func (e enclaveItem) println() {
	fmt.Println(e.Type, e.Md5, e.Shasum)
}

// Get the state of the enclave when the program starts.
func init() {
	Scan()
	// This copy only happens once.
	for k, v := range scannedEnclave {
		stableEnclave[k] = v
		stableList = append(stableList,
			scannedEnclave[k].Name + "." + scannedEnclave[k].Type)
	}
}

// Scan scans the Infinigon SGX enclave binaries.
func Scan() {
	path := ig.Env("IGBIN")
	err := filepath.Walk(path, walk)
	if err != nil {
		log.Fatal(err)
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
	name := fileInfo.Name()
	plainName := name

	switch {
	case mode.IsRegular():
		newMd5 := func(p string) string {
			data, err2 := ioutil.ReadFile(p)
			if err2 != nil {
				log.Panic(err2)
			}
			return fmt.Sprintf("%x", md5.Sum(data))
		}(_path)
		name += ".f"
		scannedEnclave[name] =
			enclaveItem{Name: plainName, Path: _path, Type: "f", Md5: newMd5, Shasum: shasum}
		scannedList = append(scannedList, name)
	case mode.IsDir():
		name += ".d"
		scannedEnclave[name] =
			enclaveItem{Name: plainName, Path: _path, Type: "d", Md5: "", Shasum: shasum}
		scannedList = append(scannedList, name)
	default:
		name += ".u"
		scannedEnclave[name] =
			enclaveItem{Name: plainName, Path: _path, Type: "u", Md5: "", Shasum: shasum}
		scannedList = append(scannedList, name)
	}

	err_ = nil
	return
}

/*************
	Errors
**************/

// EnclaveError is used to express enclave errors.
type enclaveError struct {
	What string
}

func (e enclaveError) Error() string {
	t := time.Now()
	err := fmt.Sprintf("%v: %v", time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		t.Location()),
		e.What)
	return err
}

// IsValid determines if a scanned directory matches a valid one.
func IsValid() (err_ error) {
	err_ = nil
	diff := len(stableEnclave) - len(scannedEnclave)
	switch {
	case diff > 0:
		return enclaveError{"IG17-SGX ENCLAVE: File removed!"}
	case diff < 0:
		return enclaveError{"IG17-SGX ENCLAVE: Rogue file added!"}
	default:
		return checkFileStatus()
	}
}

func checkFileStatus() (err_ error) {
	err_ = nil
	for _, x := range scannedList {
		if scannedEnclave[x].Type == "f" {
			if scannedEnclave[x].Md5 != stableEnclave[x].Md5 {
				msg := fmt.Sprintf("IG17-SGX ENCLAVE: %s chksum failed!",
					scannedEnclave[x].Name)
				return enclaveError{msg}

			}
		}
	}
	for _, x := range scannedList {
		if scannedEnclave[x].Shasum != stableEnclave[x].Shasum {
			if scannedEnclave[x].Type == "f" {
				msg := fmt.Sprintf("IG17-SGX ENCLAVE: file changed from %s to %s!",
					stableEnclave[x].Name, scannedEnclave[x].Name)
				return enclaveError{msg}
			} else {
				msg := fmt.Sprintf("IG17-SGX ENCLAVE: directory changed from %s to %s!",
					stableEnclave[x].Name, scannedEnclave[x].Name)
				return enclaveError{msg}
			}
		}
	}
	return
}

// Reset the scannedEnclave to nil before the next run.
func Reset() {
	scannedEnclave = enclaveMap{}
	scannedList = nil
}

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
	Name   string
	Path   string
	Type   string
	Md5    string
	Shasum string
}

type enclaveMap map[string]enclaveItem

var stableEnclave = enclaveMap{}
var stableList []string = nil

var scannedEnclave = enclaveMap{}
var scannedList []string = nil
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
		if oneTime {
			stableEnclave[name] =
				enclaveItem{Name: plainName, Path: _path, Type: "f", Md5: newMd5, Shasum: shasum}
			stableList = append(stableList, name)
		} else {
			scannedEnclave[name] =
				enclaveItem{Name: plainName, Path: _path, Type: "f", Md5: newMd5, Shasum: shasum}
			scannedList = append(scannedList, name)
		}
	case mode.IsDir():
		name += ".d"
		if oneTime {
			stableEnclave[name] =
				enclaveItem{Name: plainName, Path: _path, Type: "d", Md5: "", Shasum: shasum}
			stableList = append(stableList, name)
		} else {
			scannedEnclave[name] =
				enclaveItem{Name: plainName, Path: _path, Type: "d", Md5: "", Shasum: shasum}
			scannedList = append(scannedList, name)
		}
	default:
		name += ".u"
		if oneTime {
			stableEnclave[name] =
				enclaveItem{Name: plainName, Path: _path, Type: "u", Md5: "", Shasum: shasum}
			stableList = append(stableList, name)
		} else {
			scannedEnclave[name] =
				enclaveItem{Name: plainName, Path: _path, Type: "u", Md5: "", Shasum: shasum}
			scannedList = append(scannedList, name)
		}
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
	ee := enclaveError{time.Time{}, ""}
	diff := len(stableEnclave) - len(scannedEnclave)
	t := time.Now()

	switch {
	case diff > 0:
		ee = enclaveError{
			time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
				t.Location()),
			"File removed from IG17 enclave: FAIL",
		}
		return ee
	case diff < 0:
		ee = enclaveError{
			time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
				t.Location()),
			"File added to IG17 enclave: FAIL",
		}
		return ee
	default:
		return checkFileStatus()
	}
	return
}

func checkFileStatus() (err_ error) {
	err_ = nil
	t := time.Now()
	for _, x := range scannedList {
		if scannedEnclave[x].Type == "f" {
			if scannedEnclave[x].Md5 != stableEnclave[x].Md5 {
				msg := fmt.Sprintf("IG17 Enclave file %s has been tampered with: FAIL",
					scannedEnclave[x].Name)
				ee := enclaveError{
					time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
						t.Location()),
					msg,
				}
				return ee
			}
		}
	}
	for _, x := range scannedList {
		if scannedEnclave[x].Shasum != stableEnclave[x].Shasum {
			if scannedEnclave[x].Type == "f" {
				msg := fmt.Sprintf("IG17 Enclave Filename %s has been changed to %s: FAIL",
					stableEnclave[x].Name, scannedEnclave[x].Name)
				ee := enclaveError{
					time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
						t.Location()),
					msg,
				}
				return ee
			} else {
				msg := fmt.Sprintf("IG17 Enclave Directory %s has been changed to %s: FAIL",
					stableEnclave[x].Name, scannedEnclave[x].Name)
				ee := enclaveError{
					time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
						t.Location()),
					msg,
				}
				return ee
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

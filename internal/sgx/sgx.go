// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package sgx

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/donaldww/idemo2/internal/config"
	"log"
	"os"
	"path/filepath"
)

// EnclaveItem represents an artifact in the enclave.
type enclaveItem struct {
	Name   string
	Path   string
	Type   string
	Md5    string
	ShaSum string
}

type enclaveMap map[string]enclaveItem

// This is the enclave that is scanned every time.
var scannedEnclave = enclaveMap{}
var scannedList []string = nil

// This is a copy of the first scanned enclave.
var stableEnclave = enclaveMap{}
var stableList []string = nil

// println prints an enclave item.
func (e enclaveItem) println() {
	fmt.Println(e.Type, e.Md5, e.ShaSum)
}

// Get the state of the enclave when the program starts.
func init() {
	Scan()
	// This copy only happens once.
	for k, v := range scannedEnclave {
		stableEnclave[k] = v
		stableList = append(stableList,
			scannedEnclave[k].Name+"."+scannedEnclave[k].Type)
	}
}

// Scan scans the SGX enclave binaries.
func Scan() {
	path := config.Bin()
	err := filepath.Walk(path, walk)
	if err != nil {
		log.Fatal(err)
	}
}

// walk process information about each file and directory in the SGX enclave.
func walk(path string, _ os.FileInfo, _ error) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := fileInfo.Mode()
	name := fileInfo.Name()

	switch {
	case mode.IsRegular():
		key := name + ".f"
		scannedEnclave[key] =
			enclaveItem{Name: name, Path: path, Type: "f", Md5: getMd5(path), ShaSum: getShaSum(path)}
		scannedList = append(scannedList, key)
	case mode.IsDir():
		key := name + ".d"
		scannedEnclave[key] =
			enclaveItem{Name: name, Path: path, Type: "d", Md5: "", ShaSum: getShaSum(path)}
		scannedList = append(scannedList, key)
	default:
		key := name + ".u"
		scannedEnclave[key] =
			enclaveItem{Name: name, Path: path, Type: "u", Md5: "", ShaSum: getShaSum(path)}
		scannedList = append(scannedList, key)
	}

	return nil
}

// getMd5 returns the md5 encoding of a string, in hex.
func getMd5(aString string) string {
	data, err := os.ReadFile(aString)
	if err != nil {
		log.Panic(err)
	}
	m := md5.New()
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

// getShaSum returns the sha256 encoding of a string, in hex.
func getShaSum(aString string) string {
	h := sha256.New()
	h.Write([]byte(aString))
	return hex.EncodeToString(h.Sum(nil))
}

/*************
	Errors
**************/

// EnclaveError is used to express enclave errors.
type enclaveError struct {
	What string
}

func (e enclaveError) Error() string {
	return fmt.Sprintf("%v", e.What)
}

// IsValid determines if a scanned directory matches a valid one.
func IsValid() (err_ error) {
	err_ = nil
	diff := len(stableEnclave) - len(scannedEnclave)
	switch {
	case diff > 0:
		return enclaveError{"SGX SIMULATOR ENCLAVE: File removed!"}
	case diff < 0:
		return enclaveError{"SGX SIMULATOR ENCLAVE: Rogue file added!"}
	default:
		return checkFileStatus()
	}
}

func checkFileStatus() error {
	for _, x := range scannedList {
		if scannedEnclave[x].Type == "f" {
			if scannedEnclave[x].Md5 != stableEnclave[x].Md5 {
				msg := fmt.Sprintf("SGX SIMULATOR ENCLAVE: %s chksum failed!",
					scannedEnclave[x].Name)
				return enclaveError{msg}
			}
		}
	}
	for _, x := range scannedList {
		if scannedEnclave[x].ShaSum != stableEnclave[x].ShaSum {
			if scannedEnclave[x].Type == "f" {
				msg := fmt.Sprintf("SGX SIMULATOR ENCLAVE: file changed from %s to %s!",
					stableEnclave[x].Name, scannedEnclave[x].Name)
				return enclaveError{msg}
			} else {
				msg := fmt.Sprintf("SGX SIMULATOR ENCLAVE: directory changed from %s to %s!",
					stableEnclave[x].Name, scannedEnclave[x].Name)
				return enclaveError{msg}
			}
		}
	}
	return nil
}

// Reset the scannedEnclave to nil before the next run.
func Reset() {
	scannedEnclave = enclaveMap{}
	scannedList = nil
}

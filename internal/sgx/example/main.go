// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/donaldww/idemo/internal/sgx"
)

const scanInterval = 2

func sgxMain() {
	for {
		sgx.Scan()
		err := sgx.IsValid()
		if err != nil {
			fmt.Println(err)
		} else {
			t := time.Now()
			fmt.Println(time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
				t.Location()), "IG17 SGX Enclave Health: PASSED")
		}
		sgx.Reset()
		time.Sleep(scanInterval * time.Second)
	}
}

func main() {
	// stop := make(chan bool)
	// start := make(chan bool)

	go sgxMain()
	// start <- true
	time.Sleep(100 * time.Second)
	// stop <- true
}

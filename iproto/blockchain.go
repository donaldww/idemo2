// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

// Block represents each 'item' in the bc
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Blocks
var bc []Block

// bcServer handles incoming concurrent Blocks
var bcServer chan []Block
var mutex = &sync.Mutex{}
var tWindow *text.Text

// bDump logs messages into the SGX monitor widget.
func bDump(b []Block) {
	for x := range b {
		writeColorf(tWindow, cell.ColorDefault, "%v\n", x)
	}
}

func handleBlockchain(t *text.Text) {
	bcServer = make(chan []Block)
	tWindow = t

	// Create genesis block.
	tm := time.Now()
	genesisBlock := Block{0, tm.String(), 0, "", ""}
	bc = append(bc, genesisBlock)

	// Dump genesis block.
	bDump(bc)

	for {

		//TODO: add any captured transactions here
		// and generate a list of transactions.

		go handleConn()
	}
}

func handleConn() {
	go func() {
		x := newTransactions()
		for i := range x {
			newBlock, err := generateBlock(bc[len(bc)-1], i)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, bc[len(bc)-1]) {
				newBlockchain := append(bc, newBlock)
				replaceChain(newBlockchain)
			}

			bcServer <- bc
		}
	}()

	// simulate receiving broadcast
	go func() {
		for {
			time.Sleep(30 * time.Second)
			mutex.Lock()
			// output, err := json.Marshal(bc)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			mutex.Unlock()
			bDump(bc)
		}
	}()
}

func newTransactions() (t []int) {
	for x := 0; x < numberOfNodes; x++ {
		t = append(t, 3)
	}
	return
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// make sure the chain we're checking is longer than the current bc
func replaceChain(newBlocks []Block) {
	mutex.Lock()
	if len(newBlocks) > len(bc) {
		bc = newBlocks
	}
	mutex.Unlock()
}

// SHA256 hasing
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

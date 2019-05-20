// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/rand"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

// Block represents each 'item' in the bc
type Block struct {
	Index     int
	Timestamp string
	Amount    int
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Blocks
var bc []Block

var tWindow *text.Text

// bDump logs messages into the SGX monitor widget.
func bDump(b []Block) {
	tWindow.Reset()
	for x := range b {
		writeColorf(tWindow, cell.ColorDefault, "%#v\n", b[x])
	}
}

// handleBlockchain is the main point of for the blockchain window.
func handleBlockchain(t *text.Text, trig chan bool) {
	tWindow = t

	// Create genesis block.
	tm := time.Now()
	genesisBlock := Block{0, tm.String(), 0, "", ""}
	bc = append(bc, genesisBlock)

	// Dump genesis block.
	bDump(bc)

	for {
		switch {
		case <-trig:
			go handleBlocks()
		}
	}
}

func handleBlocks() {

	for _, y := range newTransactions(1) {
		newBlock, err := generateBlock(bc[len(bc)-1], y)
		if err != nil {
			log.Println(err)
			continue
		}
		if isBlockValid(newBlock, bc[len(bc)-1]) {
			newBlockchain := append(bc, newBlock)
			replaceChain(newBlockchain)
		}
	}
	bDump(bc)
}

func newTransactions(n int) []int {
	max := 5000
	result := make([]int, n)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	for i := 0; i < n; i++ {
		result[i] = r1.Intn(max) + 1
	}
	return result
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

	if len(newBlocks) > len(bc) {
		bc = newBlocks
	}
}

// SHA256 hasing
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.Amount) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, amount int) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Amount = amount
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

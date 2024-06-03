// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package blockchain

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/donaldww/idemo2/internal/term"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/nu7hatch/gouuid"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Hash                 string
	Timestamp            string
	ConsensusLeader      string
	Data                 string
	NumberOfTransactions int
	Nonce                int
	PrevHash             string
}

// Blockchain is a series of validated Blocks
var bc []Block
var tWindow *text.Text
var flag = false
var leader string

// A counter.
var __ci int
var count = func() int {
	__ci++
	return __ci
}

// bDump logs messages into the SGX monitor widget.
func bDump(b []Block) {
	i := count() - 1
	if i%9 == 0 {
		tWindow.Reset()
	}
	var color cell.Color
	if flag {
		color = cell.ColorMagenta
	} else {
		color = cell.ColorDefault
	}
	term.WriteColorf(tWindow, cell.ColorRed, " 💰")
	term.WriteColorf(tWindow, color, " %#v\n", b[i])
	flag = !flag
}

// HandleBlockchain is the main point of for the blockchain window.
func HandleBlockchain(t *text.Text, trig chan string, maxT int) {
	// tWindow is global, the program will crash if removed!
	tWindow = t
	// Create genesis block.
	genesisBlock := Block{Nonce: 0, Timestamp: time.Now().String(),
		Data: "", NumberOfTransactions: 0, Hash: "", PrevHash: "",
		ConsensusLeader: "GENESIS BLOCK"}
	bc = append(bc, genesisBlock)
	// Dump genesis block.
	bDump(bc)
	for {
		leader = <-trig
		if leader != "" {
			go handleBlocks(maxT)
		}
	}
}

func handleBlocks(trans int) {
	newBlock, err := generateBlock(bc[len(bc)-1], trans)
	if err != nil {
		panic(err)
	}
	if isBlockValid(newBlock, bc[len(bc)-1]) {
		newBlockchain := append(bc, newBlock)
		replaceChain(newBlockchain)
	}
	bDump(bc)
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Nonce+1 != newBlock.Nonce {
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

// SHA256 hashing
func calculateHash(block Block) string {
	record := string(rune(block.Nonce)) + block.Timestamp + string(rune(block.NumberOfTransactions)) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, amount int) (Block, error) {
	var newBlock Block
	t := time.Now()
	u3, err := uuid.NewV3(uuid.NamespaceURL, []byte(leader))
	if err != nil {
		panic(err)
	}
	newBlock.Nonce = oldBlock.Nonce + 1
	newBlock.Timestamp = t.String()
	newBlock.ConsensusLeader = leader
	newBlock.Data = u3.String()
	newBlock.NumberOfTransactions = amount
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)
	return newBlock, nil
}

// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Blocks
var blockchain []Block

// bcServer handles incoming concurrent Blocks
var bcServer chan []Block
var mutex = &sync.Mutex{}

// Old Main

// func main() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	bcServer = make(chan []Block)
//
// 	// create genesis block
// 	t := time.Now()
// 	genesisBlock := Block{0, t.String(), 0, "", ""}
// 	spew.Dump(genesisBlock)
// 	blockchain = append(blockchain, genesisBlock)
//
// 	httpPort := os.Getenv("PORT")
//
// 	// start TCP and serve TCP server
// 	server, err := net.Listen("tcp", ":"+httpPort)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("HTTP Server Listening on port :", httpPort)
// 	defer server.Close()
//
// 	for {
// 		conn, thisErr := server.Accept()
// 		if thisErr != nil {
// 			log.Fatal(thisErr)
// 		}
// 		go handleConn(conn)
// 	}
//
// }

func handleConn(conn net.Conn) {

	defer conn.Close()

	_, _ = io.WriteString(conn, "Enter a new BPM:")

	scanner := bufio.NewScanner(conn)

	// take in BPM from stdin and add it to blockchain after conducting necessary validation
	go func() {
		for scanner.Scan() {
			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}
			newBlock, err := generateBlock(blockchain[len(blockchain)-1], bpm)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, blockchain[len(blockchain)-1]) {
				newBlockchain := append(blockchain, newBlock)
				replaceChain(newBlockchain)
			}

			bcServer <- blockchain
			_, _ = io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()

	// simulate receiving broadcast
	go func() {
		for {
			time.Sleep(30 * time.Second)
			mutex.Lock()
			output, err := json.Marshal(blockchain)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()
			_, _ = io.WriteString(conn, string(output))
		}
	}()

	for range bcServer {
		spew.Dump(blockchain)
	}

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

// make sure the chain we're checking is longer than the current blockchain
func replaceChain(newBlocks []Block) {
	mutex.Lock()
	if len(newBlocks) > len(blockchain) {
		blockchain = newBlocks
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

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

func find(leading, trailing string, done, tick chan<- bool, stop <-chan bool) {
	leading = strings.ToLower(leading)
	trailing = strings.ToLower(trailing)

	loops := 0
	for true {
		select {
		case <-stop:
			done <- true
		default:
			key, _ := crypto.GenerateKey()
			binAddr := crypto.PubkeyToAddress(key.PublicKey)
			addr := hex.EncodeToString(binAddr[:])
			if compare(addr, leading, trailing, key) {
				done <- true
			}
		}
		loops++
		if loops > 1000 {
			loops = 0
			tick <- true
		}
	}
}

func compare(addr, leading, trailing string, key *ecdsa.PrivateKey) bool {
	found := addr[0:len(leading)] == leading && addr[len(addr)-len(trailing):len(addr)] == trailing

	if !found {
		return false
	}

	keyStr := hex.EncodeToString(crypto.FromECDSA(key))
	addrFound(addr, keyStr)
	return true
}

const numWorkers = 4
const bucketSize = 40

func main() {
	if len(os.Args) == 1 {
		errNoArg()
		os.Exit(1)
	}

	leading := os.Args[1]
	trailing := ""
	if len(os.Args) >= 3 {
		trailing = os.Args[2]
	}

	done := make(chan bool, numWorkers)
	tick := make(chan bool)
	stop := make(chan bool, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go find(leading, trailing, done, tick, stop)
	}

	loops := 0
	start := time.Now()

loop:
	for {
		select {
		case <-done:
			break loop
		case <-tick:
			fmt.Print(".")
			loops++
			if loops%bucketSize == 0 {
				diff := time.Now().Sub(start)
				rate := float64(bucketSize*1000) / diff.Seconds()
				fmt.Printf(" %8.2f addr/sec\n", rate)
				start = time.Now()
			}
		}
	}

	for j := 0; j < numWorkers; j++ {
		stop <- true
	}
	close(stop)
}

func addrFound(addr string, keyStr string) {
	println("Address found:")
	fmt.Printf("addr: 0x%s\n", addr)
	fmt.Printf("pvt: 0x%s\n", keyStr)
	println("\nexiting...")
}

func errNoArg() {
	println("You need to pass a vanity match, retry with an extra agrument like: 42 42")
	println("\nexample: go run vanieth.go 42 42")
	println("\nexiting...")
}

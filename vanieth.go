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

func find(toMatch string, done, tick chan<- bool, stop <-chan bool) {
	loops := 0
	for true {
		select {
		case <-stop:
			done <- true
		default:
			key, _ := crypto.GenerateKey()
			addr := crypto.PubkeyToAddress(key.PublicKey)
			addrStr := hex.EncodeToString(addr[:])
			if compare(addrStr, toMatch, key) {
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

func compare(addrStr string, toMatch string, key *ecdsa.PrivateKey) bool {
	toMatch = strings.ToLower(toMatch)
	addrStrMatch := strings.TrimPrefix(addrStr, toMatch)
	found := addrStrMatch != addrStr

	if !found {
		return false
	}

	keyStr := hex.EncodeToString(crypto.FromECDSA(key))
	addrFound(addrStr, keyStr)
	return true
}

const numWorkers = 4
const bucketSize = 40

func main() {
	if len(os.Args) == 1 {
		errNoArg()
		os.Exit(1)
	}

	toMatch := os.Args[1]

	done := make(chan bool, numWorkers)
	tick := make(chan bool)
	stop := make(chan bool, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go find(toMatch, done, tick, stop)
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

func addrFound(addrStr string, keyStr string) {
	println("Address found:")
	fmt.Printf("addr: 0x%s\n", addrStr)
	fmt.Printf("pvt: 0x%s\n", keyStr)
	println("\nexiting...")
}

func errNoArg() {
	println("You need to pass a vanity match, retry with an extra agrument like: 42")
	println("\nexample: go run vanieth.go 42")
	println("\nexiting...")
}

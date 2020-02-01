package main

import (
	"fmt"
	"time"
)

var peer1 *Peer
var peer2 *Peer
var peer3 *Peer
var peer4 *Peer
var peer5 *Peer
var peer6 *Peer
var peer7 *Peer
var peer8 *Peer
var peer9 *Peer
var peer10 *Peer

func main() {

	peer1 = StartPeer(&Peer{})
	peer2 = StartPeer(peer1)
	peer3 = StartPeer(peer2)
	peer4 = StartPeer(peer3)
	peer5 = StartPeer(peer4)
	peer6 = StartPeer(peer2)
	peer7 = StartPeer(peer4)
	peer8 = StartPeer(peer7)
	peer9 = StartPeer(peer3)
	peer10 = StartPeer(peer9)

	time.Sleep(200 * time.Millisecond)

	// 100 is standard
	genesisblock := createGenesisblock(100)

	peer1.setGenesisblock(genesisblock)
	peer2.setGenesisblock(genesisblock)
	peer3.setGenesisblock(genesisblock)
	peer4.setGenesisblock(genesisblock)
	peer5.setGenesisblock(genesisblock)
	peer6.setGenesisblock(genesisblock)
	peer7.setGenesisblock(genesisblock)
	peer8.setGenesisblock(genesisblock)
	peer9.setGenesisblock(genesisblock)
	peer10.setGenesisblock(genesisblock)

	// testNormal()
	// stressTest()
	for true {

	}
}

func stressTest() {
	for i := 0; i < 100; i++ {
		go peer1.Sender(peer6, 10)
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("Must receive block before printing..")
	time.Sleep(15 * time.Second)
	printLedger(peer1)
	printLedger(peer4)
}

func testNormal() {
	// Peer1 should have 999000 and peer5 100999
	peer1.Sender(peer5, 1000)

	// At least one block must win here before 10 seconds
	fmt.Println("Wait for first block ...")
	time.Sleep(10 * time.Second)
	printLedger(peer2)

	// At least one block must win here before another 10 seconds
	fmt.Println("Wait for second block ...")
	peer2.Sender(peer6, 123)
	time.Sleep(10 * time.Second)
	printLedger(peer7)

	// Wait 20 more seconds to see the entire chain for peer1
	time.Sleep(20 * time.Second)
	// Prints out the entire blockchain (Not ordered print)
	for k, v := range peer1.Blockchain {
		fmt.Println("Hash: ", k)
		fmt.Println("Parent: ", v.Parent)
		fmt.Println("Slotnr: ", v.SlotNr)
		fmt.Println()
	}

}

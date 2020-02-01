package main

import (
	"math/big"
	"net"

	accounts "./accounts"
)

// Peer struct
type Peer struct {
	Ledger    *accounts.Ledger
	PublicKey *big.Int
	SecretKey *big.Int

	// Blockchain
	Genesisblock      Genesisblock
	Currentblock      string
	Blockchain        map[string]Block
	UsedTransactions  []accounts.Transaction
	TransactionBuffer []accounts.Transaction

	// Connections
	SavedConnections []net.Conn
	SavedPorts       []string
	MessagesSent     []string
	LocalPort        string
}

// Transmission struct
type Transmission struct {
	MessageType int // 1: sending port, 2: sending list of ports
	SavedPorts  []string
	LocalPort   string
	Transaction accounts.Transaction
	ID          string
	Block       Block
}

// Genesisblock Genesis block holds information on: Seed, 10 Original peers (Public keys) and a Ledger
type Genesisblock struct {
	Seed      string
	Ledger    *accounts.Ledger
	Roundtime int64
	Starttime int64
	Hardness  *big.Int
}

// Block The block will hold a BlockID (Incremented by slotcounter). TransactionID is a list of all the transactions to be executed. Lastly it will point to the ID of its parent Block
type Block struct {
	SlotNr        int64
	Draw          *big.Int
	PublicKey     *big.Int
	TransactionID []string
	Parent        string // Change -- List of list of blocks?`? Eller hashmaps af predecessors REMEMBER TO SIGN THE WHOLE BLOCK `
	Ledger        *accounts.Ledger
	Signature     string
}

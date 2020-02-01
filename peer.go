package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	accounts "./accounts"
	keygen "./keygen"
)

const connectionsPerPeer = 10

var lastWinner int64

// StartPeer Contains the peers own server port
func StartPeer(peerConn *Peer) *Peer {

	// Reference to peer "object"
	peer := &Peer{}

	peer.SavedConnections = make([]net.Conn, 0)
	peer.SavedPorts = make([]string, 0)
	peer.MessagesSent = make([]string, 0)
	peer.Blockchain = make(map[string]Block)

	keygen.KeyGen(300)
	peer.PublicKey = keygen.GetPKInt()
	peer.SecretKey = keygen.GetSKInt()

	// Attempt to connect to ip address

	conn, err := net.Dial("tcp", "127.0.0.1:"+peerConn.LocalPort)

	isServer := true
	if err == nil {
		isServer = false
	}

	// Starts listening and creates a local port
	ln, _ := net.Listen("tcp", ":")
	peer.LocalPort = strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)

	peer.Ledger = accounts.MakeLedger()

	// Establish connection if there are no errors.
	if !isServer {
		go startUp(conn, peer)
	} else {
	}

	//go terminalInput()
	go connectionListener(isServer, ln, peer)

	return peer
}

// StartTimer https://stackoverflow.com/questions/16466320/is-there-a-way-to-do-repetitive-tasks-at-intervals
func StartTimer(peer *Peer) {

	// Time passed from GB start till now in milli
	sinceStart := float64((time.Now().UnixNano() - peer.Genesisblock.Starttime) / 1000000)
	roundTime := float64(peer.Genesisblock.Roundtime)
	round := int64(sinceStart / roundTime)

	ticker := time.NewTicker(1000 * time.Millisecond)

	lastWinner = time.Now().UnixNano()
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if peer.Ledger.Accounts[peer.PublicKey.String()] > 0 {
					DrawVar, hashedDraw := Draw(round, peer)

					isWinner := EvaluateDraw(hashedDraw, peer)
					if isWinner == true {
						newWinner := time.Now().UnixNano()
						difference := (newWinner - lastWinner) / 1000000000
						lastWinner = newWinner
						fmt.Println("Got a winner after:", difference, "seconds!")
						isWinner = false

						blockToSend := CreateWinnersBlock(peer, DrawVar, round)
						transactionsFee(peer, blockToSend)
						blockToSend.Ledger = CopyLedger(peer.Ledger)
						AddBlockToChain(blockToSend, peer)
						SendWinnerPeersBlock(blockToSend, peer)

					}
				}
				round++

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

/** -------------- SERVER PART -------------- */
/** ConnectionListener will setup the connection listener and then listen for incoming connections
*		- New (incoming) connections will be added to the list of saved connections (For two-way communication)
*		- A list of all ports saved on this peer will be sent to the connected peer
 */
func connectionListener(isServer bool, ln net.Listener, peer *Peer) {

	// Roundabout way to get the port as a string (ln.Addr().String() gives [::]:12345 and we want '12345')
	peer.LocalPort = strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	// fmt.Println("------ Local Port: " + peer.LocalPort + " ------")

	// First peer must save itself to the list
	if isServer {
		peer.SavedPorts = append(peer.SavedPorts, peer.LocalPort)
	}

	for {
		conn, _ := ln.Accept()

		// Save the incoming connection to savedConnections
		transmission := &Transmission{}
		transmission.SavedPorts = peer.SavedPorts
		transmission.MessageType = 0

		// Sends its own list of savedPorts to the newly connected peer
		enc := gob.NewEncoder(conn)
		enc.Encode(transmission)
		// 2. List has been sent
		go keepConnectionAlive(conn, peer)
	}
}

/** This method will listen (read) on all connections (individual threads) and react depending on the MessageType*/
func keepConnectionAlive(conn net.Conn, peer *Peer) {
	transmission := &Transmission{}
	isMessageSent := false

	for {
		dec := gob.NewDecoder(conn)
		err := dec.Decode(transmission)

		// Connection has been closed - Exit loop and end thread
		if err != nil {
			return
		}

		isMessageSent = false
		for _, id := range peer.MessagesSent {
			if transmission.ID == id {
				isMessageSent = true
			}
		}

		if !isMessageSent && transmission.MessageType != 0 {

			// ------------------------- CHECK IF MESSAGE IS VALID HERE -----------------------------------
			stringToVerify := transmission.ID + transmission.Transaction.To + strconv.Itoa(transmission.Transaction.Amount)
			isTransmissionAuthenticated := keygen.VerifyFromString(transmission.Transaction.Signature, transmission.Transaction.From, stringToVerify)

			peer.MessagesSent = append(peer.MessagesSent, transmission.ID)

			if transmission.MessageType == 1 {
				peer.SavedPorts = append(peer.SavedPorts, transmission.LocalPort)

				forwardMessage(*transmission, peer)
				closeAndReconnect(peer)
			}

			if transmission.MessageType == 2 && isTransmissionAuthenticated {

				// Will simply add the transaction to a buffer until it receives a block
				peer.TransactionBuffer = append(peer.TransactionBuffer, transmission.Transaction)
				time.Sleep(10 * time.Millisecond)

				forwardMessage(*transmission, peer)
			}

			if transmission.MessageType == 3 {
				blockString := BlockToString(transmission.Block)
				isBlockVerified := keygen.VerifyFromString(transmission.Block.Signature, transmission.Block.PublicKey.String(), blockString)
				allTransactionsAreValid := EvaluateBlock(transmission.Block, peer)

				if isBlockVerified && allTransactionsAreValid {
					isWinner := ReceivedBlock(transmission.Block, peer)
					if isWinner {
						forwardMessage(*transmission, peer)
						transactionsFee(peer, transmission.Block)

					}
				}
			}
		}
	}
}

func isTransactionValid(transaction accounts.Transaction, peer *Peer) bool {
	sender := transaction.From
	amount := transaction.Amount
	difference := peer.Ledger.Accounts[sender] - amount

	if amount > 0 && difference >= 0 {
		return true
	}
	return false
}

func printLedger(peer *Peer) {
	fmt.Println("----------------------------------------------------------------------------------------")
	fmt.Println()
	for k, v := range peer.Ledger.Accounts {
		fmt.Println("Peer PK: " + k)
		fmt.Print("Balance: ")
		fmt.Println(v)
		fmt.Println()
	}
	fmt.Println("----------------------------------------------------------------------------------------")
}

// Sender Sends a message to all connections in connArray */
func (s *Peer) Sender(peer *Peer, amount int) {
	transmission := &Transmission{}
	transmission.MessageType = 2
	transmission.LocalPort = ""
	transmission.ID = strconv.Itoa(time.Now().Nanosecond())

	to := peer.PublicKey.String()

	transmission.Transaction.From = s.PublicKey.String()
	transmission.Transaction.To = to
	transmission.Transaction.Amount = amount
	transmission.Transaction.ID = transmission.ID

	if isTransactionValid(transmission.Transaction, s) {
		stringToSign := transmission.ID + to + strconv.Itoa(amount)
		// ------------- Signing happens here for sender ----------------
		transmission.Transaction.Signature = keygen.SignToString(stringToSign, s.PublicKey, s.SecretKey)
		// fmt.Println(transmission.Transaction.Signature)
		// s.Ledger.Transaction(&transmission.Transaction)

		s.TransactionBuffer = append(s.TransactionBuffer, transmission.Transaction)
		// printLedger(s)

		s.MessagesSent = append(s.MessagesSent, transmission.ID)
		for _, conn := range s.SavedConnections {
			enc := gob.NewEncoder(conn)
			enc.Encode(transmission)
		}
	}
}

//setGenesisblock For every peer is the genesisblock set here
func (peer *Peer) setGenesisblock(genesisblock Genesisblock) {
	peer.Genesisblock = genesisblock
	for k, v := range genesisblock.Ledger.Accounts {
		peer.Ledger.Accounts[k] = v
	}
	go StartTimer(peer)
}

/** Closes all existing connections and establishes new connections with list */
func closeAndReconnect(peer *Peer) {
	for _, conn := range peer.SavedConnections {
		conn.Close()
	}
	connectToPeers(peer)
}

/** Sends a message to all connections in connArray */
func forwardMessage(transmission Transmission, peer *Peer) {
	for _, conn := range peer.SavedConnections {
		//fmt.Println("Forwarding message with ID: " + transmission.ID + " to " + conn.RemoteAddr().String())
		enc := gob.NewEncoder(conn)
		enc.Encode(transmission)
	}
}

/** -------------- FIRST TIME CONNECTION -------------- */
func startUp(conn net.Conn, peer *Peer) {
	// These 3 functions are executed ONCE, at the start of the peer lifecycle
	waitForPeerList(conn, peer) // 2. Connects to a port and receives list of all peers

	peer.SavedPorts = append(peer.SavedPorts, peer.LocalPort)

	connectToPeers(peer)          // 4. & 3 Connects to all ports on the list. add self
	broadcastPresence(conn, peer) // 5. Broadcast own port to all connections
}

/** This method will wait until it receives a transmission containing a list (slice) of all current ports */
func waitForPeerList(conn net.Conn, peer *Peer) {
	// Reserves space for the struct
	transmission := &Transmission{}

	// Waits and decodes the received message
	dec := gob.NewDecoder(conn)
	dec.Decode(transmission)

	// Saves the newly received ports
	peer.SavedPorts = transmission.SavedPorts
}

/** This method will dial every connection on the newly received list of ports */
func connectToPeers(peer *Peer) {
	sort.Strings(peer.SavedPorts)

	reached10Connections := false
	keepConnecting := false

	for i := 0; i < 2; i++ {
		for _, port := range peer.SavedPorts {

			// Checks if current port is own port
			isOwnPort := port == peer.LocalPort

			// If own port, start connecting
			if isOwnPort && !reached10Connections {
				keepConnecting = true
			}

			if keepConnecting && !isOwnPort {
				conn, _ := net.Dial("tcp", "127.0.0.1:"+port)
				peer.SavedConnections = append(peer.SavedConnections, conn)

				if len(peer.SavedConnections) >= connectionsPerPeer {
					reached10Connections = true
					keepConnecting = false
				}

			}

		}

		for _, conn := range peer.SavedConnections {
			// Will keep listening to the new connection
			go keepConnectionAlive(conn, peer)
		}

	}
}

/** Sends its own port to all of its new connections so they save this peers port locally */
func broadcastPresence(conn net.Conn, peer *Peer) {
	transmission := &Transmission{}
	transmission.LocalPort = peer.LocalPort
	transmission.ID = strconv.Itoa(time.Now().Nanosecond())
	transmission.MessageType = 1 // sending port

	for _, conn := range peer.SavedConnections {
		//fmt.Println(peer.SavedConnections)
		enc := gob.NewEncoder(conn)
		enc.Encode(transmission)
	}
}

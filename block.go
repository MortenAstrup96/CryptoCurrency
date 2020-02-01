package main

import (
	"encoding/gob"
	"fmt"
	"math/big"
	"strconv"
	"time"

	accounts "./accounts"
	keygen "./keygen"
)

// Draw Computes the draw where the hashed draw is a random value
func Draw(slot int64, peer *Peer) (*big.Int, *big.Int) {

	stringToSign := "lottery" + strconv.FormatInt(slot, 10) + peer.Genesisblock.Seed
	stringToSignBig := big.NewInt(0)
	stringToSignBig.SetBytes([]byte(stringToSign))

	DrawVal := keygen.SignMessage(stringToSignBig, peer.SecretKey, peer.PublicKey)

	drawValBig := big.NewInt(0)
	drawValBig.SetBytes([]byte("lottery" + peer.Genesisblock.Seed + peer.PublicKey.String() + DrawVal.String()))
	HashedVal := keygen.HashMessage(drawValBig)

	return DrawVal, HashedVal
}

//BlockToString Converts the whole block to a string
func BlockToString(block Block) string {
	longString := ""
	for _, id := range block.TransactionID {
		longString = longString + id
	}

	stringToSign := block.Draw.String() + strconv.FormatInt(block.SlotNr, 10) + block.PublicKey.String() + block.Parent + longString
	stringToSignBig := big.NewInt(0)
	stringToSignBig.SetBytes([]byte(stringToSign))

	return stringToSignBig.String()
}

// CreateWinnersBlock If a peer wins the lottery they will create a block
func CreateWinnersBlock(peer *Peer, draw *big.Int, slot int64) Block {
	block := Block{}

	block.Parent = peer.Currentblock
	block.PublicKey = peer.PublicKey
	block.SlotNr = slot
	block.Draw = draw

	for _, tx := range peer.TransactionBuffer {
		block.TransactionID = append(block.TransactionID, tx.ID)
	}

	//Sign the block
	stringToSign := BlockToString(block)
	signedBlock := keygen.SignToString(stringToSign, peer.PublicKey, peer.SecretKey)
	block.Signature = signedBlock

	return block
}

// SendWinnerPeersBlock WinnerPeer sends a transmission with the new created block
func SendWinnerPeersBlock(block Block, peer *Peer) {
	transmission := &Transmission{}
	transmission.MessageType = 3
	transmission.LocalPort = ""
	transmission.ID = strconv.Itoa(time.Now().Nanosecond())
	transmission.Block = block

	//Broadcast block
	peer.MessagesSent = append(peer.MessagesSent, transmission.ID)
	for _, conn := range peer.SavedConnections {
		enc := gob.NewEncoder(conn)
		enc.Encode(*transmission)
	}
}

func CopyLedger(ledgerToCopy *accounts.Ledger) *accounts.Ledger {
	newLedger := new(accounts.Ledger)
	newLedger = accounts.MakeLedger()
	for k, v := range ledgerToCopy.Accounts {
		newLedger.Accounts[k] = v
	}

	return newLedger
}

// ReceivedBlock Checks everything before broadcast the block
func ReceivedBlock(block Block, peer *Peer) bool {
	drawVal := block.Draw
	hashedDraw := big.NewInt(0)
	hashedDraw.SetBytes([]byte("lottery" + peer.Genesisblock.Seed + block.PublicKey.String() + drawVal.String()))
	hashedDraw = keygen.HashMessage(hashedDraw)

	//Is the draw bigger than the hardness?
	isWinner := EvaluateDraw(hashedDraw, peer)
	//Verify the draw
	if isWinner {
		verificationCheck := "lottery" + strconv.FormatInt(block.SlotNr, 10) + peer.Genesisblock.Seed
		verificationCheckBig := big.NewInt(0)
		verificationCheckBig.SetBytes([]byte(verificationCheck))

		verifiedMessage := keygen.VerifyDraw(block.Draw, verificationCheckBig, block.PublicKey)

		if verifiedMessage {
			AddBlockToChain(block, peer)
			peer.Currentblock = FindLongestChain(peer)
		}
		return verifiedMessage
	}
	return false
}

// transactionsFee When a transactions is made, the receiver of the transactions gets 1 AU less than what was sent
func transactionsFee(peer *Peer, block Block) {

	for _, id := range block.TransactionID {
		for _, tx := range peer.TransactionBuffer {

			// Found a matching pair
			if id == tx.ID {
				receiver := tx.To   //the receiver of AU
				sender := tx.From   //the sender
				amount := tx.Amount //AU coins

				//The sender sends some AU coins
				amountNowS := peer.Ledger.Accounts[sender]
				amountAfterS := amountNowS - amount
				peer.Ledger.Accounts[sender] = amountAfterS
				//fmt.Println("amount for sender: ", amountAfterS)

				//The receiver gets AU coins, but there is 1 in transactions fee
				amountNowR := peer.Ledger.Accounts[receiver]
				amountAfterR := amountNowR + amount - 1
				peer.Ledger.Accounts[receiver] = amountAfterR
				//fmt.Println("amount for receiver with fee: ", amountAfterR)
			}
		}
	}

	// Creaing a new ledger and updating it to
	for _, tx := range peer.TransactionBuffer {
		peer.UsedTransactions = append(peer.UsedTransactions, tx)
	}

	// Reset buffer
	peer.TransactionBuffer = make([]accounts.Transaction, 0)
}

// AddBlockToChain adds the block to the chain
func AddBlockToChain(block Block, peer *Peer) {
	// Convert the entire struct to a byte array to be hashed:
	blockInt := big.NewInt(0)
	blockInt = blockInt.SetBytes([]byte(fmt.Sprintf("%v", block)))
	hashedBlock := keygen.HashMessage(blockInt)

	// Sets the peers current block to the newly added block
	peer.Currentblock = hashedBlock.String()

	// Adds the block to the peers local map
	peer.Blockchain[hashedBlock.String()] = block
	//fmt.Println("New block: 	", hashedBlock)
	//fmt.Println("Parent block: 	", block.Parent)

}

// FindLongestChain calculate the longest chain
func FindLongestChain(peer *Peer) string {
	longestChain := 0
	blockHash := ""

	for k, v := range peer.Blockchain {
		count := GetNumberOfParents(v, peer)
		if count >= longestChain {
			longestChain = count
			blockHash = k
		}
	}

	return blockHash
}

// EvaluateBlock evalutes all transactions
func EvaluateBlock(block Block, peer *Peer) bool {
	foundErrors := false

	for _, id := range block.TransactionID {
		for _, tx := range peer.TransactionBuffer {

			// Found a matching pair
			if id == tx.ID {
				isValidTransaction := isBlockValid(tx, block)
				if !isValidTransaction {
					foundErrors = true
				}
			}
		}
	}

	// Returns true if no errors are found
	return !foundErrors
}

func isBlockValid(transaction accounts.Transaction, block Block) bool {
	sender := transaction.From
	amount := transaction.Amount
	difference := block.Ledger.Accounts[sender] - amount

	if amount > 0 && difference >= 0 {
		return true
	}
	return false
}

// GetNumberOfParents gets the number of the blocks before this block.
func GetNumberOfParents(block Block, peer *Peer) int {
	parent := peer.Blockchain[block.Parent]
	count := 0
	if block.Parent == "" {
		return 1
	}
	count = 1 + GetNumberOfParents(parent, peer)
	return count
}

// createGenesisblock creates the genesisblock
func createGenesisblock(difficulty int64) Genesisblock {
	genesisblock := Genesisblock{}
	genesisblock.Seed = "HelloWorldWeAreCalledNegroniFromPlanetEarth"
	genesisblock.Roundtime = 1000
	genesisblock.Starttime = time.Now().UnixNano()
	genesisblock.Hardness = CalculateHardness(difficulty)

	genesisblock.Ledger = accounts.MakeLedger()

	genesisblock.Ledger.Accounts[peer1.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer2.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer3.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer4.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer5.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer6.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer7.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer8.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer9.PublicKey.String()] = 1000000
	genesisblock.Ledger.Accounts[peer10.PublicKey.String()] = 1000000

	return genesisblock
}

// CalculateHardness Hardness is part of proof-of-stake
func CalculateHardness(difficulty int64) *big.Int {
	two := big.NewInt(2)
	twoFiveSix := big.NewInt(256)
	hardness := big.NewInt(0)

	hardness = two.Exp(two, twoFiveSix, nil) // 2^256

	delay := big.NewInt(difficulty) // (amount/TotalAmount)*(peer ratio in system) => (1/10)*(1/10) = 100

	return hardness.Div(hardness, delay) //(1/10)*(1/10)*2^256
}

// EvaluateDraw Finds out if the draw is smaller and equal than the hardness
func EvaluateDraw(draw *big.Int, peer *Peer) bool {
	return draw.Cmp(peer.Genesisblock.Hardness) <= 0 // +1 if x >  y
}

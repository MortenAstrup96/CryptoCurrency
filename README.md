# Project Overview
This project is part of the course **'Distributed Systems & Security'**. In the course we worked with the following:
* Confidentiality
* Authenticity
* Consistency
* Network & System Security
* Peer to Peer Networking

## Final Project
In our final project, my team and I created a **Distributed Ledger** using **Static Proof-of-Stake, Tree-Based, Totally Ordered Broadcasts**. The project is not a complete crypto currency, but implements many of the same features. 

## How to run
To rune the program you will need GoLang installed. To run it type the command `go run main.go peer.go block.go structs.go`. It will not look very exciting. The program will announce every time a peer wins a block. No transactions are made. The new block will build upon the previous block to create a tree. 

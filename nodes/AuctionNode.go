package main

import (
	"LogicalTime/proto"
)

type AuctionNode struct {
	id              int
	port            int
	timestamp       int
	dataConnections map[int]proto.DatabaseClient //ID to Port
}

func main() {

}

func errorHandling() { //Error occurs when contact time limit is exceeded

}

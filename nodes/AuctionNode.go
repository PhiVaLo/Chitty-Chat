package main

import (
	"LogicalTime/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strconv"
)

type AuctionNode struct {
	id              int
	port            int
	timestamp       int
	dataConnections map[int]proto.DatabaseClient //ID to Port
}

func connectToDatabase(port int) (proto.NodeClient, error) {
	// Dial the database at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", port)
	}

	return proto.NewDatabaseClient(conn), nil
}

func main() {

}

func errorHandling() { //Error occurs when contact time limit is exceeded

}

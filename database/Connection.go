package main

import (
	"LogicalTime/proto"
	"log"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectToDatabase(port int) (proto.DatabaseClient, error) {
	// Dial the database at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", port)
	}

	return proto.NewDatabaseClient(conn), nil
}

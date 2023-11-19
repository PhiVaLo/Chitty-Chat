package main

import (
	"LogicalTime/proto"
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Database struct {
	proto.UnimplementedDatabaseServer
	id           int
	port         int
	timestamp    int
	databaseConn map[int]proto.DatabaseClient //ID to Port

	currentAuction Auction
}

type Auction struct {
	id              int
	item            string
	highestBid      int
	highestBidderId int
	endTime         int // initialize when auction starts
}

var (
	port = flag.Int("port", 0, "node port number")
	id   = flag.Int("id", 0, "nodeID should be unique")
)

func main() {

	database := &Database{
		id:           *id,
		port:         *port,
		timestamp:    1,
		databaseConn: make(map[int]proto.DatabaseClient),
	}

	// Initialize database map of connected databases
	database.databaseConn[0], _ = ConnectToDatabase(5050)
	database.databaseConn[1], _ = ConnectToDatabase(5051)
	database.databaseConn[2], _ = ConnectToDatabase(5052)

	go startDatabase(database)

	// Keep the node running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) //Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel //Waits for the channel to receive a signal
}

func startDatabase(database *Database) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(database.port))

	// Error handling if node could not be created
	if err != nil {
		log.Fatalf("Could not create the node %v", err)
	}

	log.Printf("Started node at port: %d ; lamport timestamp %d \n", database.port, database.timestamp)

	proto.RegisterDatabaseServer(grpcServer, database)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (database *Database) AskForBid(ctx context.Context, in *proto.BidMessage) (*proto.AcknowledgementMessage, error) {
	/*given a bid, returns an outcome among {fail, success or exception.
	Each bid needs to be higher than the previous.
	If the auction is over it will fail, telling the node the auction is over. */

	currentState := ""

	return &proto.AcknowledgementMessage{
		Id:        int64(database.id),
		Timestamp: int64(database.timestamp),
		State:     currentState,
	}, nil
}

func (database *Database) AskForResult(ctx context.Context, in *emptypb.Empty) (*proto.ResultMessage, error) {
	//if the auction is over, it returns the result, else highest bid.

	return &proto.ResultMessage{
		Id:           int64(database.id),
		Timestamp:    int64(database.timestamp),
		ResultAmount: int64(database.currentAuction.highestBid),
		WinnerId:     int64(database.currentAuction.highestBidderId),
	}, nil
}

func (database *Database) AskForCorrespondance(ctx context.Context, in *proto.AuctionInfoMessage) (*proto.AuctionInfoMessage, error) {
	//Check info across all databases

	return &proto.AuctionInfoMessage{
		Id:              int64(database.id),
		Timestamp:       int64(database.timestamp),
		HighestBid:      int64(database.currentAuction.highestBid),
		HighestBidderId: int64(database.currentAuction.highestBidderId),
		Endtime:         int64(database.currentAuction.endTime), //May not be needed
	}, nil
}

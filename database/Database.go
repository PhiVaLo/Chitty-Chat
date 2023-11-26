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
	"sync"
	"syscall"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Database struct {
	proto.UnimplementedDatabaseServer
	id           int
	port         int
	timestamp    int
	databaseConn map[int]proto.DatabaseClient // ID to Port

	currentAuction Auction
}

type Auction struct {
	id              int
	item            string
	highestBid      int
	highestBidderId int
	endTime         int // initialize when auction starts
	hasEnded        bool
}

var (
	port = flag.Int("port", 0, "database port number")
	id   = flag.Int("id", 0, "databaseID should be unique")
)

func main() {

	flag.Parse()

	auction := &Auction{
		id:              1,
		item:            "Santa's Silver Socks",
		highestBid:      0,
		highestBidderId: -1,
		endTime:         8, // Timestamp
		hasEnded:        false,
	}

	database := &Database{
		id:             *id,
		port:           *port,
		timestamp:      1,
		databaseConn:   make(map[int]proto.DatabaseClient),
		currentAuction: *auction,
	}

	go startDatabase(database)

	// Initialize database map of connected databases
	go func() {
		database.databaseConn[1], _ = ConnectToDatabase(5050)
		database.databaseConn[2], _ = ConnectToDatabase(5051)
		database.databaseConn[3], _ = ConnectToDatabase(5052)
	}()

	go hasEnded(database)

	// Keep the database running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) // Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel // Waits for the channel to receive a signal
}

func ConnectToDatabase(port int) (proto.DatabaseClient, error) {
	// Dial the database at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", port)
	}

	return proto.NewDatabaseClient(conn), nil
}

func hasEnded(database *Database) {
	for {
		if database.currentAuction.endTime <= database.timestamp {
			database.currentAuction.hasEnded = true
			database.timestamp++
			log.Printf("Auction has ended")
			break
		}
	}
}

func startDatabase(database *Database) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(database.port))

	// Error handling if database could not be created
	if err != nil {
		log.Fatalf("Could not create the database %v", err)
	}

	log.Printf("Started database at port: %d ; lamport timestamp %d \n", database.port, database.timestamp)

	proto.RegisterDatabaseServer(grpcServer, database)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

// Update the current auction with the information from the database upon correspondance
func (auction *Auction) Update(in *proto.AuctionInfoMessage) {
	auction.highestBid = int(in.HighestBid)
	auction.highestBidderId = int(in.HighestBidderId)
	auction.endTime = int(in.Endtime)
	auction.hasEnded = in.HasEnded
}

func correspond(database *Database) {
	for id, databaseConn := range database.databaseConn {
		if id != database.id {
			var wg sync.WaitGroup // Add waitgroup

			msg := &proto.AuctionInfoMessage{
				Id:              int64(database.id),
				Timestamp:       int64(database.timestamp),
				HighestBid:      int64(database.currentAuction.highestBid),
				HighestBidderId: int64(database.currentAuction.highestBidderId),
				Endtime:         int64(database.currentAuction.endTime),
				HasEnded:        database.currentAuction.hasEnded,
			}
			var databaseCorrespondance *proto.AuctionInfoMessage
			var err error

			// If the method takes to long, go to errorhandling
			wg.Add(1)
			go func() {
				defer wg.Done()
				databaseCorrespondance, err = databaseConn.AskForCorrespondance(context.Background(), msg)

				//Checks the connection to the database
				if err != nil {
					log.Printf("Could not connect to Database #%d", id)
				} else {
					// If the database has newer information, update the current auction
					if databaseCorrespondance.State == 1 {
						database.timestamp = int(databaseCorrespondance.Timestamp)
						database.currentAuction.Update(databaseCorrespondance)
					}
				}
			}()
			wg.Wait()
		}
	}
}

func (database *Database) AskForBid(ctx context.Context, in *proto.BidMessage) (*proto.AcknowledgementMessage, error) {
	/*given a bid, returns an outcome among {fail, success or exception.
	Each bid needs to be higher than the previous.
	If the auction is over it will fail, telling the node the auction is over. */

	log.Printf("Bidder %d wants to bid %d", in.Id, in.BidAmount)

	log.Printf("Correspond with other databases")
	correspond(database)

	var currentState int

	if database.currentAuction.hasEnded { // if auction is over
		currentState = 2 //Auction has ended
		log.Printf("Currentstate was set to 2: Auction Ended")

	} else if database.currentAuction.highestBid < int(in.BidAmount) { // If bid is higher that the current bid
		currentState = 0 //Success
		log.Printf("Currentstate was set to 0: Success")

	} else if database.currentAuction.highestBid >= int(in.BidAmount) { // If bid is not higher than the current bid
		currentState = 1 //Bid too low
		log.Printf("Currentstate was set to 1: Bid is too low")
	}

	log.Printf("Currentstate is %d", currentState)

	if currentState == 0 { // Bidder is allowed to bid
		database.timestamp++

		//Update values of the auction
		database.currentAuction.highestBid = int(in.BidAmount)
		database.currentAuction.highestBidderId = int(in.Id)

		log.Printf("Currentbid was updated")
	}

	return &proto.AcknowledgementMessage{
		Id:         int64(database.id),
		Timestamp:  int64(database.timestamp),
		State:      int64(currentState),
		HighestBid: int64(database.currentAuction.highestBid),
	}, nil
}

func (database *Database) AskForResult(ctx context.Context, in *emptypb.Empty) (*proto.ResultMessage, error) {
	log.Printf("The current result was requested \n")
	return &proto.ResultMessage{
		Id:           int64(database.id),
		Timestamp:    int64(database.timestamp),
		ResultAmount: int64(database.currentAuction.highestBid),
		WinnerId:     int64(database.currentAuction.highestBidderId),
		HasEnded:     database.currentAuction.hasEnded,
	}, nil
}

func (database *Database) AskForCorrespondance(ctx context.Context, in *proto.AuctionInfoMessage) (*proto.AuctionInfoMessage, error) {
	// Check what information is the newest and return it
	if database.timestamp > int(in.Timestamp) {
		log.Printf("Database %d has newer information - Copies to %d...", database.id, in.Id)
		return &proto.AuctionInfoMessage{
			Id:              int64(database.id),
			Timestamp:       int64(database.timestamp),
			HighestBid:      int64(database.currentAuction.highestBid),
			HighestBidderId: int64(database.currentAuction.highestBidderId),
			Endtime:         int64(database.currentAuction.endTime),
			HasEnded:        database.currentAuction.hasEnded,
			State:           1,
		}, nil
	}

	return &proto.AuctionInfoMessage{
		State: 0,
	}, nil
}

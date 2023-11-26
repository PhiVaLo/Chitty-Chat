package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/protobuf/types/known/emptypb"
)

type AuctionNode struct {
	id           int
	port         int
	timestamp    int
	databaseConn map[int]proto.DatabaseClient // ID to Port
}

var (
	port = flag.Int("port", 0, "auctionNode port number")
	id   = flag.Int("id", 0, "auctionNodeID should be unique")
)

func main() {

	flag.Parse()

	auctionNode := &AuctionNode{
		id:           *id,
		port:         *port,
		timestamp:    1,
		databaseConn: make(map[int]proto.DatabaseClient),
	}

	// Initialize database map of connected databases
	auctionNode.databaseConn[0], _ = ConnectToDatabase(5050)
	auctionNode.databaseConn[1], _ = ConnectToDatabase(5051)
	auctionNode.databaseConn[2], _ = ConnectToDatabase(5052)

	// Handle
	log.Printf("To bid type: bid <amount> \n To request result type: result \n")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		userInput := scanner.Text()

		splitInput := strings.Split(userInput, " ")

		var bidAmount int
		var err error

		if len(splitInput) > 1 {
			bidAmount, err = strconv.Atoi(splitInput[1])
		}

		switch {
		case splitInput[0] == "bid" && err == nil && bidAmount > 0: // If the user makes a bid as input, they are bidding on the auction
			auctionNode.BidOnAuction(bidAmount)

		case userInput == "result": // If the user gives result as input, they are requesting the current result
			auctionNode.Result()

		default:
			log.Printf("Input not recognized \n  To bid type: bid <amount> \n To request result type: result \n")
		}

	}

	// Keep the auctionNode running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) // Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel // Waits for the channel to receive a signal
}

func (auctionNode *AuctionNode) BidOnAuction(bid int) {

	var wg sync.WaitGroup // Add waitgroup

	msg := &proto.BidMessage{
		Id:        int64(auctionNode.id),
		Timestamp: int64(auctionNode.timestamp),
		BidAmount: int64(bid),
	}

	acknowledgementList := make(map[int]*proto.AcknowledgementMessage)

	for _, database := range auctionNode.databaseConn {

		// If the method takes to long, go to errorhandling
		wg.Add(1)
		go func() {
			defer wg.Done()
			acknowledgementMessage, err := database.AskForBid(context.Background(), msg)

			if err != nil {
				log.Fatalf("Error message: %v", err)
			}

			acknowledgementList[int(acknowledgementMessage.Id)] = acknowledgementMessage
		}()
		//go errorHandling(&wg)
		wg.Wait()
	}

	var newestMessage *proto.AcknowledgementMessage
	highestTimestamp := math.MinInt64

	for _, result := range acknowledgementList {
		if int(result.Timestamp) > highestTimestamp {

			newestMessage = result
			highestTimestamp = int(result.Timestamp)
		}
	}

	if newestMessage.State == 0 { // Bidder was allowed to bid
		auctionNode.timestamp++
		log.Printf("You bidded %d \n", bid)

	} else if newestMessage.State == 1 { //Bid is too low
		log.Printf("Bid is too low. Currentbid is: %d", newestMessage.HighestBid)

	} else if newestMessage.State == 2 { //Auction is over
		log.Printf("Auction is over - you can no longer bid")
	}

}

func ConnectToDatabase(port int) (proto.DatabaseClient, error) {
	// Dial the database at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", port)
	}

	return proto.NewDatabaseClient(conn), nil
}

func (auctionNode *AuctionNode) Result() {
	log.Printf("You requested the current result \n")
	var wg sync.WaitGroup // Add waitgroup

	resultList := make(map[int]*proto.ResultMessage)

	// Get resultmessages from all databases
	for _, database := range auctionNode.databaseConn {

		// If the method takes to long, go to errorhandling
		wg.Add(1)

		go func() {
			defer wg.Done()
			resultMessage, err := database.AskForResult(context.Background(), &emptypb.Empty{})

			if err != nil {
				log.Fatalf("This is an error: %v", err)
				wg.Done()
				return
			}

			resultList[int(resultMessage.Id)] = resultMessage
		}()
		//go errorHandling(&wg)
		wg.Wait()
	}

	// Check if all results are the same, if not, choose the result with the highest timestamp
	var newestMessage *proto.ResultMessage
	highestTimestamp := math.MinInt64

	for _, result := range resultList {
		if int(result.Timestamp) > highestTimestamp {

			newestMessage = result
			highestTimestamp = int(result.Timestamp)
		}
	}

	// If auction is not over (Winnerid haven't been set)
	if !newestMessage.HasEnded {
		log.Printf("The current highest bid is: %d", newestMessage.ResultAmount)
	} else { // If auction is over (WinnerId has a value)
		log.Printf("The auction is over, the winner is Bidder %d with the highest bid: %d", newestMessage.WinnerId, newestMessage.ResultAmount)
	}
}

func errorHandling(wg *sync.WaitGroup) { // Error occurs when contact time limit is exceeded
	// Count to 20 seconds to give method time to finish
	duration := 20 * time.Second
	timer := time.After(duration)

	select { //Sælcat
	case <-timer:
		// Code to execute after 20 seconds
		wg.Done()
	}
}

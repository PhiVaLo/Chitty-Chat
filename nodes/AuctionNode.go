package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type AuctionNode struct {
	id           int
	port         int
	timestamp    int
	databaseConn map[int]proto.DatabaseClient //ID to Port
}

var (
	port = flag.Int("port", 0, "auctionNode port number")
	id   = flag.Int("id", 0, "auctionNodeID should be unique")
)

func main() {

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

		bid := "bid"
		bidAmount, err := strconv.Atoi(splitInput[1])

		switch {
		case bid == "bid" && err != nil && bidAmount > 0: //If the user a bid as input, they are bidding on the auction
			log.Printf("You bidded %d \n", bidAmount)

		case userInput == "result": //If the user gives result as input, they are requesting the current result
			log.Printf("You requested the current result \n") //CALL THE RESULT THING

		default:
			log.Printf("Input not recognized \n  To bid type: bid <amount> \n To request result type: result \n")
		}

	}

	// Keep the auctionNode running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) //Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel //Waits for the channel to receive a signal
}

func (auctionNode *AuctionNode) BidOnAuction(bid int) {
	auctionNode.timestamp++

	var wg sync.WaitGroup //Add waitgroup

	msg := &proto.BidMessage{
		Id:        int64(auctionNode.id),
		Timestamp: int64(auctionNode.timestamp),
		BidAmount: int64(bid),
	}

	for _, database := range auctionNode.databaseConn {

		go func() {
			defer wg.Done()
			acknowledgementMessage, _ := database.AskForBid(context.Background(), msg)

		}()
		wg.Wait()
	}
}

func errorHandling() { //Error occurs when contact time limit is exceeded

}

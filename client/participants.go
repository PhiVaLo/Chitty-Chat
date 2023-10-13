package main

import (
	"PhysicalTime/proto"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	id         int
	portNumber int
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	client := &Client{
		id:         1,
		portNumber: *clientPort,
	}
	serverConnection, _ := connectToServer()

	// Wait for the client (user) to ask for the time
	go waitForTimeRequest(client)

	go waitForClientJoin(client)
	go waitForClientLeave(client)

	// Keep the client running
	for {
	}
}

func connectToServer() (proto.TimeAskClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	} else {
		log.Printf("Connected to the server at port %d\n", *serverPort)
	}
	return proto.NewTimeAskClient(conn), nil
}

func participant(){

}

func leaveChat(){

}

func joinChat(){

}

func receiveMessage(){
	//PRINT (Client received the message: *** at lamport timestamp)
}

func sendMessage(client *Client, serverConnection *connectToServer){
	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		message := fmt.Sprintf("Participant sends the message: %s\n", input); //Skal ikke printes i log, skal sendes til server

		// Ask the server for the time
		timeReturnMessage, err := serverConnection.AskForTime(context.Background(), &proto.PublishMessage{
			ClientId: int64(client.id),
			message: input,
		})

		if err != nil {
			log.Printf(err.Error())
		} else {
			message = fmt.Sprintf(message + " at the time %s\n", timeReturnMessage.ServerName, timeReturnMessage.Time);
			return message
			//message := fmt.Sprintf("%s at the time %s\n", timeReturnMessage.ServerName, timeReturnMessage.Time);
		}
	}
}




func waitForTimeRequest(client *Client) {
	// Connect to the server
	serverConnection, _ := connectToServer()

	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		log.Printf("Client asked for time with input: %s\n", input)

		// Ask the server for the time
		timeReturnMessage, err := serverConnection.AskForTime(context.Background(), &proto.AskForTimeMessage{
			ClientId: int64(client.id),
		})

		if err != nil {
			log.Printf(err.Error())
		} else {
			log.Printf("Server %s says the time is %s\n", timeReturnMessage.ServerName, timeReturnMessage.Time)
		}
	}
}
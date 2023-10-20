package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"sync"
)

var globalId int
var globalMutex sync.Mutex

func setGlobalId(id int) {
    globalMutex.Lock()
    globalId = id
    globalMutex.Unlock()
}

func getGlobalId() int {
	globalMutex.Lock()
    id := globalId
    globalMutex.Unlock()
    return id
}

type Client struct {
	id         int
	portNumber int
	timestamp  int
	serverConn proto.PublishClient
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Each client has a unique id
	setGlobalId(getGlobalId()+1)

	// Create a client
	client := &Client{
		id:         getGlobalId(),
		portNumber: *clientPort,
		timestamp:  0,
	}

	// Wait for the client (user) to send message
	go sendMessage(client)

	// Keep the client running
	for {
	}
}

func connectToServer(client *Client) (proto.PublishClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	}

	return proto.NewPublishClient(conn), nil
}

func sendMessage(client *Client) {
	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)

	//Connects to server
	serverConnection, _ := connectToServer(client)
	
	serverConnection.AskToJoin(context.Background(), &proto.JoinOrLeaveMessage{
		ClientId:  int64(client.id),
		Timestamp: int64(client.timestamp),
		Port:      int64(client.portNumber),
	})
	log.Printf("Participant %d joined the Chitty-chat server at port %d\n", client.id, *serverPort)

	client.timestamp++ //They joined, so the timestamp is updated

	for scanner.Scan() {
		input := scanner.Text()

		if len(input) > 128 {
			log.Printf("Message is too big (maximum length 128 characters)")
			continue //Starts the for loop from scratch
		}

		//Sends message timestamp++
		client.timestamp++
		log.Printf("Participant sends the message: %s at the time %d \n", input, client.timestamp)

		// Ask the server to publish the message
		/*broadcastReturnMessage, err :=*/ serverConnection.AskForPublish(context.Background(), &proto.PublishMessage{
			ClientId:  int64(client.id),
			Timestamp: int64(client.timestamp),
			Message:   input,
		})
		/*
		if err != nil {
			log.Printf(err.Error())
		}

		//Client receives a message, so the timestamp is updated
		if client.timestamp < int(broadcastReturnMessage.Timestamp) {
			client.timestamp = int(broadcastReturnMessage.Timestamp)
		}
		client.timestamp++*/
	}
}

func (client *Client) AskForMessageBroadcast(ctx context.Context, in *proto.PublishMessage) (*proto.BroadcastMessage, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Participant %d send the message: %s at lamport timestamp %d \n", in.ClientId, in.Message, client.timestamp)

	return nil, nil
}

func (client *Client) AskForJoinBroadcast(ctx context.Context, in *proto.JoinOrLeaveMessage) (*proto.BroadcastMessage, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Participant %d joined at lamport timestamp %d \n", in.ClientId, client.timestamp)

	return nil, nil
}

func (client *Client) AskForLeaveBroadcast(ctx context.Context, in *proto.JoinOrLeaveMessage) (*proto.BroadcastMessage, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Participant %d left at lamport timestamp %d \n", in.ClientId, client.timestamp)

	return nil, nil
}

/*
func receiveMessage(client *Client, serverConnection proto.PublishClient) {
	//PRINT (Client received the message: *** at lamport timestamp)
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
*/

package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	proto.UnimplementedBroadcastServer
	id         int
	portNumber int
	timestamp  int
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
	clientID   = flag.Int("id", 0, "clientID should be unique")
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	client := &Client{
		id:         *clientID,
		portNumber: *clientPort,
		timestamp:  1,
	}

	// Starts client as a grpc server for listening capabilities
	go startClient(client)

	//Connects to server
	serverConnection, _ := connectToServer(client)

	// Wait for the client (user) to send message
	go sendMessage(client, serverConnection)

	// Keep the client running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) //Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel //Waits for the channel to receive a signal

	//Asks the server to leave
	serverConnection.AskToLeave(context.Background(), &proto.JoinOrLeaveMessage{
		ClientId:  int64(client.id),
		Timestamp: int64(client.timestamp),
		Port:      int64(client.portNumber),
	})
}

func startClient(client *Client) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(client.portNumber))

	// Error handling if client could not be created
	if err != nil {
		log.Fatalf("Could not create the client %v", err)
	}

	log.Printf("Started client at port: %d ; lamport timestamp %d \n", client.portNumber, client.timestamp)

	// Register the grpc server and serve its listener
	proto.RegisterBroadcastServer(grpcServer, client)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
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

func sendMessage(client *Client, serverConnection proto.PublishClient) {
	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)

	//Connects to server
	//serverConnection, _ := connectToServer(client)

	serverConnection.AskToJoin(context.Background(), &proto.JoinOrLeaveMessage{
		ClientId:  int64(client.id),
		Timestamp: int64(client.timestamp),
		Port:      int64(client.portNumber),
	})

	client.timestamp++ //They joined, so the timestamp is updated
	log.Printf("Participant %d joined the Chitty-chat server at port %d at timestamp %d \n", client.id, *serverPort, client.timestamp)

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
		/*broadcastReturnMessage, err :=*/
		serverConnection.AskForPublish(context.Background(), &proto.PublishMessage{
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

func (client *Client) AskForMessageBroadcast(ctx context.Context, in *proto.PublishMessage) (*emptypb.Empty, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Server sends: Participant %d send the message: %s at lamport timestamp %d \n", in.ClientId, in.Message, client.timestamp)

	return &emptypb.Empty{}, nil
}

func (client *Client) AskForJoinBroadcast(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Server sends: Participant %d joined at lamport timestamp %d \n", in.ClientId, client.timestamp)

	return &emptypb.Empty{}, nil
}

func (client *Client) AskForLeaveBroadcast(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Server sends: Participant %d left at lamport timestamp %d \n", in.ClientId, client.timestamp)

	return &emptypb.Empty{}, nil
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

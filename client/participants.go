package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"google.golang.org/protobuf/types/known/emptypb"

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

	var wg sync.WaitGroup //Add waitgroup

	// Starts client as a grpc server for listening capabilities
	wg.Add(1)
	go startClient(client, &wg)
	wg.Wait()

	//Connects to server
	serverConnection, _ := connectToServer()

	// Wait for the client (user) to send message
	go sendMessage(client, serverConnection)

	// Keep the client running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) //Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel //Waits for the channel to receive a signal

	//Asks the server to leave
	client.timestamp++
	log.Printf("This client left the server at lamport timestamp %d \n", client.timestamp)

	serverConnection.AskToLeave(context.Background(), &proto.JoinOrLeaveMessage{
		ClientId:  int64(client.id),
		Timestamp: int64(client.timestamp),
		Port:      int64(client.portNumber),
	})
}

func startClient(client *Client, wg *sync.WaitGroup) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(client.portNumber))

	// Error handling if client could not be created
	if err != nil {
		log.Fatalf("Could not create the client %v", err)
	}

	log.Printf("Started client at port: %d ; lamport timestamp %d \n", client.portNumber, client.timestamp)
	wg.Done()

	// Register the grpc server and serve its listener
	proto.RegisterBroadcastServer(grpcServer, client)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func connectToServer() (proto.PublishClient, error) {
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

	client.timestamp++
	log.Printf("Client #%d asks to join the server at lamport timestamp %d \n", client.id, client.timestamp)

	serverConnection.AskToJoin(context.Background(), &proto.JoinOrLeaveMessage{
		ClientId:  int64(client.id),
		Timestamp: int64(client.timestamp),
		Port:      int64(client.portNumber),
	})

	for scanner.Scan() {
		input := scanner.Text()

		if len(input) > 128 {
			log.Printf("Message is too big (maximum length 128 characters)")
			continue //Starts the for loop from scratch
		}

		//Sends message timestamp++
		client.timestamp++
		log.Printf("Participant sends the message: %s - at lamport timestamp %d \n", input, client.timestamp)

		// Ask the server to publish the message
		serverConnection.AskForPublish(context.Background(), &proto.PublishMessage{
			ClientId:  int64(client.id),
			Timestamp: int64(client.timestamp),
			Message:   input,
		})
	}
}

func (client *Client) AskForMessageBroadcast(ctx context.Context, in *proto.PublishMessage) (*emptypb.Empty, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Client #%d sends the message: %s - at lamport timestamp %d \n", in.ClientId, in.Message, client.timestamp)

	return &emptypb.Empty{}, nil
}

func (client *Client) AskForJoinBroadcast(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Client #%d joined at lamport timestamp %d \n", in.ClientId, client.timestamp)

	return &emptypb.Empty{}, nil
}

func (client *Client) AskForLeaveBroadcast(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Client receives the message therefore timestamp++
	if client.timestamp < int(in.Timestamp) {
		client.timestamp = int(in.Timestamp)
	}
	client.timestamp++

	log.Printf("Client #%d left the server at lamport timestamp %d \n", in.ClientId, client.timestamp)

	return &emptypb.Empty{}, nil
}

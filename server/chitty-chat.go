package main

import (
	"LogicalTime/proto"
	"context"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

// Server Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedPublishServer
	name      string
	port      int
	timestamp int
	clients   map[int]proto.PublishClient
}

// Used to get the user-defined port for the server from the command line
var port = flag.Int("port", 0, "server port number")

func main() {
	// Get the port from the command line when the server is run
	flag.Parse()

	// Create a server struct
	server := &Server{
		name:      "Chitty-Chat",
		port:      *port,
		timestamp: 0,
	}

	// Start the server
	go startServer(server)

	// Keep the server running
	for {
	}
}

func startServer(server *Server) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	// Error handling if server could not be created
	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	proto.RegisterPublishServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (server *Server) AddClientToServer(clientID int, client proto.PublishClient) {
	server.clients[clientID] = client
}

func (server *Server) RemoveClientFromServer(clientID int) {
	server.clients[clientID] = nil
}

func (server *Server) AskForPublish(ctx context.Context, in *proto.PublishMessage) (*proto.BroadcastMessage, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	//PRINT (Participant x send the message: *** at lamport timestamp)
	log.Printf("Participant %d send the message: %s at lamport timestamp %d \n", in.ClientId, in.Message, server.timestamp)

	//Broadcast to the rest of the participants
	//TODO: Broadcast statement to all clients except the one who originally send it
	for id, client := range server.clients {
		if id != int(in.ClientId) && client != nil {
			server.timestamp++
			msg := &proto.BroadcastMessage{
				Timestamp: int64(server.timestamp),
				Message:   in.Message,
			}

			//client.BroadcastMessage(context.Background(), msg)
		}
	}

	//Sends statement to original client
	server.timestamp++
	return &proto.BroadcastMessage{
		Timestamp: int64(server.timestamp),
		Message:   in.Message,
	}, nil
}

/*
	func (server *Server) participantLeft(ctx context.Context, in *proto.PublishMessage) (*proto.TimeMessage, error){
		//Server receives the message therefore timestamp++
		if(server.timestamp < in.timestamp) {
			server.timestamp = in.timestamp
		}
		server.timestamp++;


		//PRINT (Participant X left Chitty-Chat at Lamport time L)
		log.Printf("Participant %d left Chitty-chat at timestamp %s\n", in.ClientId)
	}

	func (server *Server) participantJoined(ctx context.Context, in *proto.PublishMessage) (*proto.TimeMessage, error){
		//Server receives the message therefore timestamp++
		if(server.timestamp < in.timestamp) {
			server.timestamp = in.timestamp
		}
		server.timestamp++;


		//PRIT (Participant X  joined Chitty-Chat at Lamport time L)
		log.Printf("Participant %d joined Chitty-chat at timestamp %s\n", in.ClientId, )
	}
*/

/*
func (server *Server) test(ctx context.Context, in *proto.PublishMessage) (*proto.BroadcastMessage, error) {
	log.Printf("Server: Participant %d send the message: %s at lamport timestamp %d \n", in.ClientId, in.Message, server.timestamp)
	return &proto.BroadcastMessage{
		Timestamp: int64(server.timestamp),
		Message:   in.Message,
	}, nil
}
*/

package main

import (
	"LogicalTime/proto"
	"context"
	"flag"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc/credentials/insecure"

	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	// "fmt"
)

// Server Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedPublishServer
	name      string
	port      int
	timestamp int
	clients   map[int]int
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
		timestamp: 1,
		clients:   make(map[int]int),
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

	log.Printf("Started server at port: %d ; lamport timestamp %d \n", server.port, server.timestamp)

	// Register the grpc server and serve its listener
	proto.RegisterPublishServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func connectToClient(port int) (proto.BroadcastClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", port)
	}

	return proto.NewBroadcastClient(conn), nil
}

func (server *Server) AddClientToServer(clientId int, clientPort int) {
	//Checks if the client port is already added
	exists := server.clients[clientId]

	//Adds the clientport if it does not already exist
	if exists == 0 {
		server.clients[clientId] = clientPort
	} else {
		log.Fatalf("Client with id %d already exists", clientId)
	}
}

func (server *Server) RemoveClientFromServer(clientId int) {
	// remove clientID from the map
	delete(server.clients, clientId)
}

func (server *Server) AskForPublish(ctx context.Context, in *proto.PublishMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	//Server prints in terminal for us to see
	log.Printf("Participant %d sends the message: %s - at lamport timestamp %d \n", in.ClientId, in.Message, server.timestamp)

	var wg sync.WaitGroup //Add waitgroup
	for id, port := range server.clients {
		if port != 0 {
			wg.Add(1)
			clientConn, _ := connectToClient(port)
			server.timestamp++ //Timestamp go up for every event (so for every client its send to)

			go func() {
				msg := &proto.PublishMessage{ //We only send the information, not the entire string
					ClientId:  int64(in.ClientId),
					Timestamp: int64(server.timestamp),
					Message:   in.Message,
				}

				log.Printf("Server broadcasts message to Client #%d at lamport timestamp %d \n", id, server.timestamp)
				clientConn.AskForMessageBroadcast(context.Background(), msg)
				defer wg.Done()
			}()
		}
		wg.Wait()
	}
	return &emptypb.Empty{}, nil
}

func (server *Server) AskToJoin(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	log.Printf("Client #%d joined at lamport timestamp %d", in.ClientId, server.timestamp)

	//Adds the client connection to the map
	server.AddClientToServer(int(in.ClientId), int(in.Port))

	//Broadcast the joining participant to all existing clients
	var wg sync.WaitGroup //Add waitgroup
	for id, port := range server.clients {
		if port != 0 {
			wg.Add(1)
			clientConn, _ := connectToClient(port)
			server.timestamp++ //Timestamp go up for every event (so for every client it is send to)

			go func() {
				msg := &proto.JoinOrLeaveMessage{ //We only send the information, not the entire string
					ClientId:  int64(in.ClientId),
					Timestamp: int64(server.timestamp),
				}

				log.Printf("Server sends join message to Client #%d at lamport timestamp %d \n", id, server.timestamp)
				clientConn.AskForJoinBroadcast(context.Background(), msg)
				defer wg.Done()
			}()
		}
		wg.Wait()
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) AskToLeave(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	log.Printf("Client #%d left at lamport timestamp %d", in.ClientId, server.timestamp)

	//Remove client from the server
	server.RemoveClientFromServer(int(in.ClientId))

	if len(server.clients) < 2 { //If list of clients only contain the leaving participant or less
		return nil, nil
	} else {
		//Send message to all remaining clients
		var wg sync.WaitGroup //Add waitgroup
		for id, port := range server.clients {
			if port != 0 {
				wg.Add(1)
				clientConn, _ := connectToClient(port)
				server.timestamp++ //Timestamp go up for every event (so for every client it is send to)

				go func() {
					msg := &proto.JoinOrLeaveMessage{ //We only send the information, not the entire string
						ClientId:  int64(in.ClientId),
						Timestamp: int64(server.timestamp),
					}

					log.Printf("Server sends leave message to Client #%d at lamport timestamp %d \n", id, server.timestamp)
					clientConn.AskForLeaveBroadcast(context.Background(), msg)
					defer wg.Done()
				}()
			}
			wg.Wait()
		}
	}
	return &emptypb.Empty{}, nil
}

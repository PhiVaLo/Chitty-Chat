package main

import (
	"LogicalTime/proto"
	"context"
	"flag"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
	// "fmt"
)

// Server Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedPublishServer
	name      string
	port      int
	timestamp int
	clients   []int
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
		clients:   make([]int, 0),
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

func connectToClient(port int) (proto.BroadcastClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", port)
	}

	return proto.NewBroadcastClient(conn), nil
}

func (server *Server) AddClientToServer(clientPort int) {
	//Checks if the client port is already added
	exists := false
	for _, port := range server.clients {
		if port == clientPort {
			exists = true
		}
	}

	//Adds the clientport if it does not already exist
	if !exists {
		server.clients = append(server.clients, clientPort)
	} else {
		log.Fatalf("Client with port %d already exists", clientPort)
	}
}

func (server *Server) RemoveClientFromServer(clientPort int) {
	// remove clientID from the map
	for index, port := range server.clients {
		if port == clientPort {
			server.clients[index] = -1
		}
	}
}

func (server *Server) AskForPublish(ctx context.Context, in *proto.PublishMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	//Server prints in terminal for us to see
	log.Printf("Participant %d send the message: %s at lamport timestamp %d \n", in.ClientId, in.Message, server.timestamp)

	var wg sync.WaitGroup //Add waitgroup
	for _, port := range server.clients {
		if port > 0 /*&& id != int(in.ClientId)*/ {
			server.timestamp++ //Timestamp go up for every event (so for every client its send to)

			msg := &proto.PublishMessage{ //We only send the information, not the entire string
				ClientId:  int64(in.ClientId),
				Timestamp: int64(server.timestamp),
				Message:   in.Message,
			}

			wg.Add(1)

			clientConn, _ := connectToClient(port)

			go func(c proto.BroadcastClient) {
				c.AskForMessageBroadcast(context.Background(), msg)
				defer wg.Done()
			}(clientConn)
		}
	}
	wg.Wait()

	/* Remove return
	//Sends message to the original client
	server.timestamp++ //Timestamp goes up, since server creates an event
	return &proto.BroadcastMessage{
		Timestamp: int64(server.timestamp),
		Message:   in.Message,
	}, nil*/
	return &emptypb.Empty{}, nil
}

func (server *Server) AskToJoin(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	//Adds the client connection to the map
	server.AddClientToServer(int(in.Port))

	//Broadcast the joining participant to all existing clients
	var wg sync.WaitGroup //Add waitgroup
	for _, port := range server.clients {
		if port > 0 /*&& id != int(in.ClientId)*/ {
			server.timestamp++ //Timestamp go up for every event (so for every client it is send to)

			msg := &proto.JoinOrLeaveMessage{ //We only send the information, not the entire string
				ClientId:  int64(in.ClientId),
				Timestamp: int64(server.timestamp),
			}

			wg.Add(1)

			clientConn, _ := connectToClient(port)

			go func(c proto.BroadcastClient) {
				c.AskForJoinBroadcast(context.Background(), msg)
				defer wg.Done()
			}(clientConn)
		}
	}
	wg.Wait()

	return &emptypb.Empty{}, nil
}

func (server *Server) AskToLeave(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	if len(server.clients) < 2 { //If list of clients only contain the leaving participant or less
		return nil, nil
	} else {
		//Send message to all remaining clients
		var wg sync.WaitGroup //Add waitgroup
		for _, port := range server.clients {
			if port > 0 /*&& id != int(in.ClientId)*/ {
				server.timestamp++ //Timestamp go up for every event (so for every client it is send to)

				msg := &proto.JoinOrLeaveMessage{ //We only send the information, not the entire string
					ClientId:  int64(in.ClientId),
					Timestamp: int64(server.timestamp),
				}

				wg.Add(1)

				clientConn, _ := connectToClient(port)

				go func(c proto.BroadcastClient) {
					c.AskForLeaveBroadcast(context.Background(), msg)
					defer wg.Done()
				}(clientConn)
			}
		}
		wg.Wait()
	}

	//Remove client from the server
	server.RemoveClientFromServer(int(in.Port))

	return &emptypb.Empty{}, nil
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

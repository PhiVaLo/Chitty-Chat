package main

import (
	"LogicalTime/proto"
	"context"
	"flag"
	"log"
	"net"
	"os/exec"
	"strconv"

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
	clients   map[int]*miniClient
}

type miniClient struct {
	id    int
	port  int
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
		clients: make(map[int]*miniClient),
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

func (server *Server) AddClientToServer(client *miniClient) {
	server.clients[client.id] = client
}

func (server *Server) RemoveClientFromServer(clientID int) {
	// remove clientID from the map
	delete(server.clients, clientID)
}
func (server *Server) BroadcastMessageToClients(message *proto.PublishMessage, ignoreId int) {
    for id, client := range server.clients {
        if client != nil && client.id != ignoreId {
            go func(c *miniClient) {
                c.AskForBroadcast(context.Background(), message)

            }(client)
        }
    }
}


func (server *Server) AskForPublish(ctx context.Context, in *proto.PublishMessage) (*proto.BroadcastMessage, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	//Server prints in terminal for us to see
	log.Printf("Participant %d send the message: %s at lamport timestamp %d \n", in.ClientId, in.Message, server.timestamp)

	//Sends the message to the rest of the participants, except the one who originally send it
	for id, client := range server.clients {
		if id != int(in.ClientId) && client != nil {
			server.timestamp++ //Timestamp go up for every event (so for every client its send to)
			msg := &proto.PublishMessage{ //We only send the information, not the entire string
				ClientId:  int64(in.ClientId),
				Timestamp: int64(server.timestamp),
				Message:   in.Message,
			}
			// Skal omskrives da clients map er en ministruct
			client.AskForBroadcast(context.Background(), msg) //Sends the information to the clients
		}
	}

	//Sends message to the original client
	server.timestamp++ //Timestamp goes up, since server creates an event
	return &proto.BroadcastMessage{
		Timestamp: int64(server.timestamp),
		Message:   in.Message,
	}, nil
}

func (server *Server) AskToJoin(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Broadcast the joining participant to all existing participants

	//Add client to server (map)
	miniclient := &miniClient{
		id:     int(in.ClientId),
		port:   int(in.Port),
	}
	
	server.AddClientToServer(miniclient);

	//Client aleady prints the joined message to itself

	return nil, nil
}

func (server *Server) AskToLeave(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error){
	server.timestamp++; //Server recieves that a client leaves
	
	if(server.clients == nil){ //If list of clients is empty
		return nil, nil
	} else {
		//Send message to all remaining participants
		//"Participant x left the server at lamport timestamp: x"
	}
	
	//Remove client from the server
	server.RemoveClientFromServer(int(in.ClientId))
	
	return nil, nil
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

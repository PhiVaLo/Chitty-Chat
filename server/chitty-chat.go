package main

import (
	"LogicalTime/proto"
	"context"
	"flag"
	"log"
	"net"
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
	clients   map[int]proto.PublishClient
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
		clients: make(map[int]proto.PublishClient),
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

func (server *Server) AddClientToServer(clientId int, client proto.PublishClient) {
	server.clients[clientId] = client
}

func (server *Server) RemoveClientFromServer(clientID int) {
	// remove clientID from the map
	delete(server.clients, clientID)
}

func (server *Server) AskForPublish(ctx context.Context, in *proto.PublishMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	//Server prints in terminal for us to see
	log.Printf("Participant %d send the message: %s at lamport timestamp %d \n", in.ClientId, in.Message, server.timestamp)

	for _, client := range server.clients {
        if client != nil /*&& id != int(in.ClientId)*/ {
			server.timestamp++ //Timestamp go up for every event (so for every client its send to)

			msg := &proto.PublishMessage{ //We only send the information, not the entire string
				ClientId:  int64(in.ClientId),
				Timestamp: int64(server.timestamp),
				Message:   in.Message,
			}
			
			//Add waitgroup
            go func(c proto.PublishClient) {
                c.AskForMessageBroadcast(context.Background(), msg)
            }(client)
        }
    }
	/* Remove return
	//Sends message to the original client
	server.timestamp++ //Timestamp goes up, since server creates an event
	return &proto.BroadcastMessage{
		Timestamp: int64(server.timestamp),
		Message:   in.Message,
	}, nil*/
	return nil, nil
}

func (server *Server) AskToJoin(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error) {
	//Server receives the message therefore timestamp++
	if server.timestamp < int(in.Timestamp) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++

	//Creates a connection to the client
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(in.Port)), grpc.WithInsecure())
    if err != nil {
		log.Fatalf("Could not connect to client: %v", err)
    }
	clientConnection := proto.NewPublishClient(conn)
	
	//Adds the client connection to the map
	server.AddClientToServer(int(in.ClientId), clientConnection);

	//Broadcast the joining participant to all existing clients
	for _, client := range server.clients {
        if client != nil /*&& id != int(in.ClientId)*/ {
			server.timestamp++ //Timestamp go up for every event (so for every client it is send to)

			msg := &proto.JoinOrLeaveMessage{ //We only send the information, not the entire string
				ClientId:  int64(in.ClientId),
				Timestamp: int64(server.timestamp),
			}
			
			//Add waitgroup
            go func(c proto.PublishClient) {
                c.AskForJoinBroadcast(context.Background(), msg)
            }(client)
        }
    }

	return nil, nil
}

func (server *Server) AskToLeave(ctx context.Context, in *proto.JoinOrLeaveMessage) (*emptypb.Empty, error){
	//Server receives the message therefore timestamp++
	if(server.timestamp < int(in.Timestamp)) {
		server.timestamp = int(in.Timestamp)
	}
	server.timestamp++;
	
	if(len(server.clients) < 2){ //If list of clients only contain the leaving participant or less
		return nil, nil
	} else {
		//Send message to all remaining clients		
		for _, client := range server.clients {
			if client != nil /*&& id != int(in.ClientId)*/ {
				server.timestamp++ //Timestamp go up for every event (so for every client it is send to)
	
				msg := &proto.JoinOrLeaveMessage{ //We only send the information, not the entire string
					ClientId:  int64(in.ClientId),
					Timestamp: int64(server.timestamp),
				}
				
				//Add waitgroup
				go func(c proto.PublishClient) {
					c.AskForLeaveBroadcast(context.Background(), msg)
				}(client)
			}
		}
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

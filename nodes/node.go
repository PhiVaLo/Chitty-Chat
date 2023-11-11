package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type Node struct {
	proto.UnimplementedNodeServer
	id        int
	port      int
	timestamp int
	requestQ  map[int]int //ID to Timestamp
	replyQ    map[int]int //ID to Timestamp
	inCS      bool
	nodes     map[int]int //ID to Port
}

var (
	port = flag.Int("port", 0, "node port number")
	id   = flag.Int("id", 0, "nodeID should be unique")
)

func main() {
	// Get the port and id from the command line when the node is initialized
	flag.Parse()

	// Create a node struct
	node := &Node{
		id:        *id,
		port:      *port,
		timestamp: 1,
		requestQ:  make(map[int]int),
		replyQ:    make(map[int]int),
		inCS:      false,
		nodes:     make(map[int]int),
	}

	// Starts node as a grpc server - Listens for requests
	go startNode(node)

	// Keep the node running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) //Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel //Waits for the channel to receive a signal
}

func startNode(node *Node) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(node.port))

	// Error handling if client could not be created
	if err != nil {
		log.Fatalf("Could not create the node %v", err)
	}

	log.Printf("Started node at port: %d ; lamport timestamp %d \n", node.port, node.timestamp)

	proto.RegisterNodeServer(grpcServer, node)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func sendPermissionMessage(node *Node) {
	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		node.timestamp++
		log.Printf("Node #%d asks for permission at lamport timestamp %d \n", node.id, node.timestamp)

		//Needs to do this to all connections (connection proto.NodeClient)
		connection.AskForPermission(context.Background(), &proto.PermissionMessage{
			NodeId:     int64(node.id),
			Timestamp:  int64(node.timestamp),
			Permission: node.inCS,
		})
	}
}

func (node *Node) AskForPermission(ctx context.Context, in *proto.PermissionMessage) (*proto.PermissionMessage, error) {
	//Client receives the message therefore timestamp++
	if node.timestamp < int(in.Timestamp) {
		node.timestamp = int(in.Timestamp)
	}
	node.timestamp++

	log.Printf("Node #%d asks for permission - at lamport timestamp %d \n", in.NodeId, node.timestamp)

	//Implementation

	return &proto.PermissionMessage{
		NodeId:     int64(node.id),
		Timestamp:  int64(node.timestamp),
		Permission: !node.inCS,
	}, nil
}

func (node *Node) NotifyExit(ctx context.Context, in *proto.PermissionMessage) (*proto.PermissionMessage, error) {
	//Client receives the message therefore timestamp++
	if node.timestamp < int(in.Timestamp) {
		node.timestamp = int(in.Timestamp)
	}
	node.timestamp++

	log.Printf("Node #%d notifies exit of CS - at lamport timestamp %d \n", in.NodeId, node.timestamp)

	//Implementation

	return &proto.PermissionMessage{
		NodeId:     int64(node.id),
		Timestamp:  int64(node.timestamp),
		Permission: node.inCS,
	}, nil
}

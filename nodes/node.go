package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	proto.UnimplementedNodeServer
	id        int
	port      int
	timestamp int
	requestQ  map[int]int //ID to Timestamp
	replyQ    map[int]int //ID to Timestamp
	inCS      bool
	waiting   bool
	nodes     map[int]proto.NodeClient //ID to Port
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
		waiting:   false,
		nodes:     make(map[int]proto.NodeClient),
	}

	//Initialize nodes map of connected nodes
	node.nodes[1], _ = connectToNode(5050)
	node.nodes[2], _ = connectToNode(5051)
	node.nodes[3], _ = connectToNode(5052)

	// Starts node as a grpc server - Listens for requests
	go startNode(node)

	go sendPermissionMessage(node)

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

	// Error handling if node could not be created
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
	// Wait for input in the node terminal
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		node.timestamp++
		log.Printf("Node #%d asks for permission at lamport timestamp %d \n", node.id, node.timestamp)

		//Add our own to requestq
		node.requestQ[node.id] = node.timestamp
		node.waiting = true

		var wg sync.WaitGroup //Add waitgroup
		for id, nodeConn := range node.nodes {
			if id != node.id {
				wg.Add(1)
				node.timestamp++ //Timestamp go up for every event (so for every node it is send to)

				go func() {
					msg := &proto.PermissionMessage{
						NodeId:     int64(node.id),
						Timestamp:  int64(node.timestamp),
						Permission: node.inCS,
					}

					log.Printf("Node #%d asks for permission of Node #%d to enter CS at lamport timestamp %d \n", node.id, id, node.timestamp)
					reply, _ := nodeConn.AskForPermission(context.Background(), msg)

					// Use reply from AskForPermission
					node.replyQ[int(reply.NodeId)] = int(reply.Timestamp)

					//Node receives the request therefore timestamp++
					if node.timestamp < int(reply.Timestamp) {
						node.timestamp = int(reply.Timestamp)
					}
					node.timestamp++
					log.Printf("Node #%d receives a reply from node #%d at lamport timestamp %d \n", node.id, reply.NodeId, node.timestamp)

					defer wg.Done()
				}()
			}
			wg.Wait()
		}

		if len(node.replyQ) != len(node.nodes)-1 {
			log.Fatalf("Something went wrong before entering CS for Node #%d", node.id)
		}

		//Enter CS - set inCS to true and waiting to false
		node.inCS = true
		node.waiting = false
		node.timestamp++
		log.Printf("Node #%d has entered CS at lamport timestamp %d \n", node.id, node.timestamp)

		//Delete node from requestQ and Clear replyQ
		delete(node.requestQ, node.id)
		node.replyQ = make(map[int]int)

		//Exit CS
		node.inCS = false
		node.timestamp++
		log.Printf("Node #%d has exited CS at lamport timestamp %d \n", node.id, node.timestamp)
	}
}

func connectToNode(port int) (proto.NodeClient, error) {
	// Dial the node at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", port)
	}

	return proto.NewNodeClient(conn), nil
}

func (node *Node) AskForPermission(ctx context.Context, in *proto.PermissionMessage) (*proto.PermissionMessage, error) {
	//Node receives the message therefore timestamp++
	if node.timestamp < int(in.Timestamp) {
		node.timestamp = int(in.Timestamp)
	}
	node.timestamp++

	log.Printf("Node #%d asks for permission to CS - at lamport timestamp %d \n", in.NodeId, node.timestamp)

	//Implementation
	node.requestQ[int(in.NodeId)] = int(in.Timestamp) //Adds asking node to receiving nodes requestQ

	if len(node.requestQ) > 1 {
		SmallestTimestamp := math.MaxInt64
		SmallestTimestampId := math.MaxInt64

		for id, timestamp := range node.requestQ {
			if timestamp < SmallestTimestamp { //If timestamp is smaller than the smallest timestamp
				SmallestTimestamp = timestamp
				SmallestTimestampId = id
			}
		}

		if SmallestTimestampId != int(in.NodeId) { //If not equal the requesting node wait til node is out of CS to continue
			//Wait til SmallestTimestampId is out of CS
			for node.inCS || node.waiting { //While the node is in CS or waiting for replies
				//Wait
			}
		}

		//If the smallest timestamp is the requesting node, continue
	}

	//If only the requesting node is in the queue, return reply to asking node
	return &proto.PermissionMessage{
		NodeId:     int64(node.id),
		Timestamp:  int64(node.timestamp),
		Permission: node.inCS,
	}, nil
}

func (node *Node) NotifyExit(ctx context.Context, in *proto.PermissionMessage) (*proto.PermissionMessage, error) {
	//Node receives the message therefore timestamp++
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

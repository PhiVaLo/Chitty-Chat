package main

import (
	"LogicalTime/proto"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

type Node struct {
	proto.UnimplementedNodeServer
	id        int
	port      int
	timestamp int
	nodes     map[int]int
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
		nodes:     make(map[int]int),
	}

	// Starts node as a grpc server
	go startNode(node)

	// Keep the client running / Wait for shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM) //Notifies the channel when ctrl+c is pressed or the terminal is interrupted

	<-signalChannel //Waits for the channel to receive a signal
}

func startNode(node *Node) {

}

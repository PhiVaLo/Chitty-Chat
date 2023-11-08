package main

import (
	"LogicalTime/proto"
	"flag"
)

type Node struct {
	proto.UnimplementedNodeServer
	id        int
	port      int
	timestamp int
	nodes     map[int]int
}

var (
	nodePort = flag.Int("port", 0, "node port number")
	nodeID   = flag.Int("id", 0, "nodeID should be unique")
)

func main() {

}

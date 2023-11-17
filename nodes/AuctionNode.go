package main

import "LogicalTime/proto"

type AuctionNode struct {
	proto.UnimplementedAuctionNodeServer
	id              int
	port            int
	timestamp       int
	requestQ        map[int]int              //ID to Timestamp
	replyQ          map[int]int              //ID to Timestamp
	nodeConnections map[int]proto.NodeClient //ID to Port
}

func main() {

}

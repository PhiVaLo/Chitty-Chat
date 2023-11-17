package main

import "LogicalTime/proto"

type Database struct {
	proto.UnimplementedDatabaseServer
	id           int
	port         int
	timestamp    int
	databaseConn map[int]proto.DatabaseClient //ID to Port
}

func main() {

}

# From root-directory



Terminal-1 (server):

    go run ./server/chitty-chat.go -port 5454

Terminal-2 (client):

    go run ./client/participants.go -port 8080 -sPort 5454


Generate proto files:

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/proto.proto
```
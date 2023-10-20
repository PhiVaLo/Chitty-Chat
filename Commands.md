# From root-directory


Terminal-1 (server):

```bash
go run ./server/chitty-chat.go -port 5454
```

Terminal-N (client):

```bash
go run ./client/participants.go -cPort 8080 -sPort 5454
go run ./client/participants.go -cPort 8081 -sPort 5454
go run ./client/participants.go -cPort 8082 -sPort 5454
```

Generate proto files:

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/proto.proto
```









Co-authored-by: Patrick Søborg <ptso@itu.dk>
Co-authored-by: Anna Høybye Johansen <annaj@itu.dk>
From root-directory:

Terminal-1: 

    go run ./server/chitty-chat.go -port 5454

Terminal-2: 

    go run ./client/participants.go -port 8080 -sPort 5454
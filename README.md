# How to run the program
The program is relatively simple to run. First and foremost the server needs to be initialized through
the following command (assuming it is run from the root):

```bash
go run ./server/chitty-chat.go -port 5454
```

As shown the server needs to initialized to a port which does not necessarily have to be 5454 but it is
used in this example. Afterwards one can initialize a client through the following command:

```bash
go run ./client/participants.go -cPort 8080 -sPort 5454 -id 1
```

This command needs a client port (-cPort 8080), a server port (-sPort 5454) which needs to match up with port given to the server in the first command and an id (-id 1). From here on out the client is able to speak with the server: simply type whatever message in the client you wish to send and it
will be send to the server. It is also possible to open multiple clients as long as their ports and ids are unique and they reference the same server port.

## Note:
* The server port should always stay the same.
* The client ports needs to be unique.
* The client ids needs to be unique.
* Both ports and ids are integers


--------------------------------------------------------------------

### Below are the full commands used for this project in the report

From root-directory

Terminal-1 (server):

```bash
go run ./server/chitty-chat.go -port 5454
```

Terminal-N (client):

```bash
go run ./client/participants.go -cPort 8080 -sPort 5454 -id 1
go run ./client/participants.go -cPort 8081 -sPort 5454 -id 2
go run ./client/participants.go -cPort 8082 -sPort 5454 -id 3
```

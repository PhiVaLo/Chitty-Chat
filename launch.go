package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

func main() {
	//Starts all of the databases
	go startDatabase(5050, 1)
	go startDatabase(5051, 2)
	go startDatabase(5052, 3)

	//Starts all of the nodes
	go startAuctionNode(8080, 1)
	go startAuctionNode(8081, 2)
}

func startDatabase(port int, id int) {
	var cmd *exec.Cmd

	command := fmt.Sprintf("go run ./database/Database.go -port %d -id %d", port, id)

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("x-terminal-emulator", "-e", command)
	case "windows":
		cmd = exec.Command("cmd", "/c", command)
	default:
		log.Fatalf("System not supported: Could not open terminal window")
	}

	cmd.Start()
}

func startAuctionNode(port int, id int) {
	var cmd *exec.Cmd

	command := fmt.Sprintf("go run ./nodes/AuctionNode.go -port %d -id %d", port, id)

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("x-terminal-emulator", "-e", command)
	case "windows":
		cmd = exec.Command("cmd", "/c", command)
	default:
		log.Fatalf("System not supported: Could not open terminal window")
	}

	cmd.Start()
}

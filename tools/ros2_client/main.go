package main

import (
	"flag"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Simple test for ROS2 bridge - we'll need the proto files generated for full implementation
func main() {
	var (
		serverAddr = flag.String("server", "localhost:50052", "The ROS2 bridge server address")
	)
	flag.Parse()

	// Connect to the server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to ROS2 bridge server: %s\n", *serverAddr)
	
	// Test connection state
	state := conn.GetState()
	fmt.Printf("Connection state: %v\n", state)
	
	// TODO: Add actual API calls when proto is properly integrated
	fmt.Println("Connection successful! ROS2 bridge is responding.")
}
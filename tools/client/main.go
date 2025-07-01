package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	systemv1 "github.com/proto_api/services/system-manager/gen/system/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		serverAddr   = flag.String("server", "localhost:50051", "The server address")
		command      = flag.String("cmd", "reboot", "Command to execute: reboot, shutdown, drs-stop, drs-restart, drs-status, recorder-stop, recorder-restart, recorder-status, disk")
		delaySeconds = flag.Int("delay", 0, "Delay in seconds for reboot/shutdown")
		diskPath     = flag.String("path", "/", "Path for disk usage check")
	)
	flag.Parse()

	// Connect to the server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := systemv1.NewSystemServiceClient(conn)

	fmt.Printf("Connected to server: %s\n", *serverAddr)
	fmt.Printf("Executing command: %s\n", *command)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute command based on input
	switch strings.ToLower(*command) {
	case "reboot":
		executeReboot(ctx, client, *delaySeconds)
	case "shutdown":
		executeShutdown(ctx, client, *delaySeconds)
	case "drs-stop":
		executeDrsAction(ctx, client, systemv1.ManageDrsServiceRequest_SERVICE_ACTION_STOP)
	case "drs-restart":
		executeDrsAction(ctx, client, systemv1.ManageDrsServiceRequest_SERVICE_ACTION_RESTART)
	case "drs-status":
		executeDrsAction(ctx, client, systemv1.ManageDrsServiceRequest_SERVICE_ACTION_STATUS)
	case "recorder-stop":
		executeRecorderAction(ctx, client, systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_STOP)
	case "recorder-restart":
		executeRecorderAction(ctx, client, systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_RESTART)
	case "recorder-status":
		executeRecorderAction(ctx, client, systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_STATUS)
	case "disk":
		executeDiskUsage(ctx, client, *diskPath)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		fmt.Println("Available commands: reboot, shutdown, drs-stop, drs-restart, drs-status, recorder-stop, recorder-restart, recorder-status, disk")
	}
}

func executeReboot(ctx context.Context, client systemv1.SystemServiceClient, delaySeconds int) {
	fmt.Printf("Requesting reboot with %d seconds delay\n", delaySeconds)
	
	req := &systemv1.RebootRequest{
		DelaySeconds: int32(delaySeconds),
	}

	resp, err := client.Reboot(ctx, req)
	if err != nil {
		log.Fatalf("Reboot request failed: %v", err)
	}

	fmt.Printf("Reboot response:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
}

func executeShutdown(ctx context.Context, client systemv1.SystemServiceClient, delaySeconds int) {
	fmt.Printf("Requesting shutdown with %d seconds delay\n", delaySeconds)
	
	req := &systemv1.ShutdownRequest{
		DelaySeconds: int32(delaySeconds),
	}

	resp, err := client.Shutdown(ctx, req)
	if err != nil {
		log.Fatalf("Shutdown request failed: %v", err)
	}

	fmt.Printf("Shutdown response:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
}

func executeDrsAction(ctx context.Context, client systemv1.SystemServiceClient, action systemv1.ManageDrsServiceRequest_ServiceAction) {
	fmt.Printf("Managing DRS service: %v\n", action)
	
	req := &systemv1.ManageDrsServiceRequest{
		Action: action,
	}

	resp, err := client.ManageDrsService(ctx, req)
	if err != nil {
		log.Fatalf("DRS service request failed: %v", err)
	}

	fmt.Printf("DRS service response:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
	if resp.ServiceStatus != "" {
		fmt.Printf("  Status: %s\n", resp.ServiceStatus)
	}
}

func executeRecorderAction(ctx context.Context, client systemv1.SystemServiceClient, action systemv1.ManageRecorderServiceRequest_ServiceAction) {
	fmt.Printf("Managing Recorder service: %v\n", action)
	
	req := &systemv1.ManageRecorderServiceRequest{
		Action: action,
	}

	resp, err := client.ManageRecorderService(ctx, req)
	if err != nil {
		log.Fatalf("Recorder service request failed: %v", err)
	}

	fmt.Printf("Recorder service response:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
	if resp.ServiceStatus != "" {
		fmt.Printf("  Status: %s\n", resp.ServiceStatus)
	}
}

func executeDiskUsage(ctx context.Context, client systemv1.SystemServiceClient, path string) {
	if path == "/" {
		fmt.Printf("Getting disk usage for default path\n")
	} else {
		fmt.Printf("Getting disk usage for: %s\n", path)
	}
	
	req := &systemv1.GetDiskUsageRequest{
		Path: path,
	}

	resp, err := client.GetDiskUsage(ctx, req)
	if err != nil {
		log.Fatalf("Disk usage request failed: %v", err)
	}

	fmt.Printf("Disk usage response:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
	
	if resp.DiskUsage != nil {
		usage := resp.DiskUsage
		fmt.Printf("  Filesystem: %s\n", usage.Filesystem)
		fmt.Printf("  Total: %.2f GB\n", float64(usage.TotalBytes)/1024/1024/1024)
		fmt.Printf("  Used: %.2f GB\n", float64(usage.UsedBytes)/1024/1024/1024)
		fmt.Printf("  Free: %.2f GB\n", float64(usage.FreeBytes)/1024/1024/1024)
		fmt.Printf("  Usage: %.1f%%\n", usage.UsagePercentage)
	}
}
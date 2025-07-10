package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	systemv1 "github.com/drs-api/services/module-agent/gen/system/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		serverAddr    = flag.String("server", "localhost:50051", "The server address")
		command       = flag.String("cmd", "reboot", "Command to execute: reboot, shutdown, service, list-services, disk, ptp, ptp-all")
		serviceName   = flag.String("service", "", "Service name for service management")
		serviceAction = flag.String("action", "status", "Service action: start, stop, restart, status, enable, disable")
		delaySeconds  = flag.Int("delay", 0, "Delay in seconds for reboot/shutdown")
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
	case "service":
		if *serviceName == "" {
			fmt.Println("Error: -service flag is required for service command")
			return
		}
		executeServiceAction(ctx, client, *serviceName, *serviceAction)
	case "list-services":
		executeListServices(ctx, client)
	case "disk":
		executeDiskUsage(ctx, client)
	case "ptp":
		executePTPCheck(ctx, client, false)
	case "ptp-all":
		executePTPCheck(ctx, client, true)
	// Backward compatibility shortcuts
	case "drs-stop":
		executeServiceAction(ctx, client, "services/drs_sensor", "stop")
	case "drs-restart":
		executeServiceAction(ctx, client, "services/drs_sensor", "restart")
	case "drs-status":
		executeServiceAction(ctx, client, "services/drs_sensor", "status")
	case "recorder-stop":
		executeServiceAction(ctx, client, "services/drs_recorder", "stop")
	case "recorder-restart":
		executeServiceAction(ctx, client, "services/drs_recorder", "restart")
	case "recorder-status":
		executeServiceAction(ctx, client, "services/drs_recorder", "status")
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		fmt.Println("Available commands: reboot, shutdown, service, list-services, disk, ptp, ptp-all")
		fmt.Println("  service examples:")
		fmt.Println("    -cmd=service -service=services/drs_sensor -action=stop")
		fmt.Println("    -cmd=service -service=services/drs_recorder -action=restart")
		fmt.Println("    -cmd=service -service=drs_sensor.service -action=stop (legacy format)")
		fmt.Println("    -cmd=service -service=drs_recorder.service -action=restart (legacy format)")
		fmt.Println("  For backward compatibility: drs-stop, drs-restart, drs-status, recorder-stop, recorder-restart, recorder-status")
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

func executeServiceAction(ctx context.Context, client systemv1.SystemServiceClient, serviceName, actionStr string) {
	fmt.Printf("Managing service: %s, action: %s\n", serviceName, actionStr)
	
	// Convert systemd service name to resource name
	resourceName := convertSystemdNameToResourceName(serviceName)
	
	// Execute the appropriate method based on action
	switch strings.ToLower(actionStr) {
	case "start":
		req := &systemv1.StartServiceRequest{Name: resourceName}
		resp, err := client.StartService(ctx, req)
		if err != nil {
			log.Fatalf("Start service request failed: %v", err)
		}
		printServiceResult("Start", resp.Service)
		
	case "stop":
		req := &systemv1.StopServiceRequest{Name: resourceName}
		resp, err := client.StopService(ctx, req)
		if err != nil {
			log.Fatalf("Stop service request failed: %v", err)
		}
		printServiceResult("Stop", resp.Service)
		
	case "restart":
		req := &systemv1.RestartServiceRequest{Name: resourceName}
		resp, err := client.RestartService(ctx, req)
		if err != nil {
			log.Fatalf("Restart service request failed: %v", err)
		}
		printServiceResult("Restart", resp.Service)
		
	case "status":
		req := &systemv1.GetServiceRequest{Name: resourceName}
		resp, err := client.GetService(ctx, req)
		if err != nil {
			log.Fatalf("Get service request failed: %v", err)
		}
		printServiceResult("Status", resp)
		
	case "enable":
		req := &systemv1.EnableServiceRequest{Name: resourceName}
		resp, err := client.EnableService(ctx, req)
		if err != nil {
			log.Fatalf("Enable service request failed: %v", err)
		}
		printServiceResult("Enable", resp.Service)
		
	case "disable":
		req := &systemv1.DisableServiceRequest{Name: resourceName}
		resp, err := client.DisableService(ctx, req)
		if err != nil {
			log.Fatalf("Disable service request failed: %v", err)
		}
		printServiceResult("Disable", resp.Service)
		
	default:
		fmt.Printf("Invalid action: %s\n", actionStr)
		fmt.Println("Valid actions: start, stop, restart, status, enable, disable")
		return
	}
}

// convertSystemdNameToResourceName converts systemd service name to resource name
func convertSystemdNameToResourceName(systemdName string) string {
	// Remove .service suffix if present
	serviceName := strings.TrimSuffix(systemdName, ".service")
	
	// Map common systemd names to resource names
	switch serviceName {
	case "drs_sensor":
		return "services/drs_sensor"
	case "drs_recorder":
		return "services/drs_recorder"
	default:
		// If it already looks like a resource name, return as-is
		if strings.Contains(serviceName, "/") {
			return serviceName
		}
		// Otherwise, assume it's a service resource
		return "services/" + serviceName
	}
}

// printServiceResult prints the service information
func printServiceResult(action string, service *systemv1.Service) {
	if service == nil {
		fmt.Printf("%s operation completed but no service info returned\n", action)
		return
	}
	
	fmt.Printf("%s operation completed successfully\n", action)
	fmt.Printf("Service Info:\n")
	fmt.Printf("  Name: %s\n", service.Name)
	fmt.Printf("  State: %s\n", service.State.String())
	fmt.Printf("  Enabled: %v\n", service.Enabled)
	if service.Description != "" {
		fmt.Printf("  Description: %s\n", service.Description)
	}
	if service.UptimeSeconds > 0 {
		fmt.Printf("  Uptime: %d seconds (%.1f minutes)\n", service.UptimeSeconds, float64(service.UptimeSeconds)/60)
	}
}

func executeListServices(ctx context.Context, client systemv1.SystemServiceClient) {
	fmt.Println("Listing services...")
	
	req := &systemv1.ListServicesRequest{}

	resp, err := client.ListServices(ctx, req)
	if err != nil {
		log.Fatalf("List services request failed: %v", err)
	}

	fmt.Printf("List services response:\n")
	
	if len(resp.Services) > 0 {
		fmt.Printf("\nServices:\n")
		for i, service := range resp.Services {
			fmt.Printf("  [%d] %s\n", i+1, service.Name)
			fmt.Printf("      State: %s\n", service.State.String())
			fmt.Printf("      Enabled: %v\n", service.Enabled)
			if service.Description != "" {
				fmt.Printf("      Description: %s\n", service.Description)
			}
			if service.UptimeSeconds > 0 {
				fmt.Printf("      Uptime: %d seconds (%.1f minutes)\n", service.UptimeSeconds, float64(service.UptimeSeconds)/60)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("  No services found")
	}
}

func executeDiskUsage(ctx context.Context, client systemv1.SystemServiceClient) {
	fmt.Printf("Getting disk usage for primary disk\n")
	
	req := &systemv1.GetDiskUsageRequest{}

	resp, err := client.GetDiskUsage(ctx, req)
	if err != nil {
		log.Fatalf("Disk usage request failed: %v", err)
	}

	fmt.Printf("Disk usage response:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
	
	if resp.DiskUsage != nil {
		usage := resp.DiskUsage
		fmt.Printf("  Total: %.2f GB\n", float64(usage.TotalBytes)/1024/1024/1024)
		fmt.Printf("  Used: %.2f GB\n", float64(usage.UsedBytes)/1024/1024/1024)
		fmt.Printf("  Free: %.2f GB\n", float64(usage.FreeBytes)/1024/1024/1024)
		fmt.Printf("  Usage: %.1f%%\n", usage.UsagePercentage)
	}
}

func executePTPCheck(ctx context.Context, client systemv1.SystemServiceClient, includeRemote bool) {
	if includeRemote {
		fmt.Println("Checking PTP sync status for local and remote devices...")
	} else {
		fmt.Println("Checking local PTP sync status...")
	}
	
	req := &systemv1.GetPTPStatusRequest{
		IncludeRemoteDevices: includeRemote,
	}
	
	resp, err := client.GetPTPStatus(ctx, req)
	if err != nil {
		log.Fatalf("PTP check request failed: %v", err)
	}
	
	fmt.Printf("PTP sync response:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
	
	if resp.Success && resp.LocalStatus != nil {
		fmt.Printf("\n[Local PTP Status]\n")
		printPTPStatus(resp.LocalStatus)
		
		if includeRemote && len(resp.RemoteStatuses) > 0 {
			fmt.Printf("\n[Remote Devices]\n")
			for _, remote := range resp.RemoteStatuses {
				fmt.Printf("\n  Device: %s (%s)\n", remote.DeviceName, remote.IpAddress)
				fmt.Printf("  Reachable: %v\n", remote.IsReachable)
				if remote.IsReachable && remote.Status != nil {
					printPTPStatus(remote.Status)
				} else if !remote.IsReachable && remote.ErrorMessage != "" {
					fmt.Printf("  Error: %s\n", remote.ErrorMessage)
				}
			}
		}
	}
}

func printPTPStatus(status *systemv1.PTPStatus) {
	fmt.Printf("  Clock ID: %s\n", status.ClockId)
	fmt.Printf("  Master Offset: %d ns (%.3f ms)\n", status.MasterOffsetNs, float64(status.MasterOffsetNs)/1000000.0)
	fmt.Printf("  Ingress Time: %d\n", status.IngressTime)
	fmt.Printf("  GM Present: %v\n", status.GmPresent)
	fmt.Printf("  GM Identity: %s\n", status.GmIdentity)
	fmt.Printf("  Is Synced: %v\n", status.IsSynced)
}
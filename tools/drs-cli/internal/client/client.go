package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	modulev1 "github.com/drs-api/tools/drs-cli/gen/drs/module/v1"
	ros2bridgev1 "github.com/drs-api/tools/drs-cli/gen/drs/ros2bridge/v1"
)

type Client struct {
	conn               *grpc.ClientConn
	serviceManager     modulev1.ServiceManagerServiceClient
	systemControl      modulev1.SystemControlServiceClient
	monitoring         modulev1.MonitoringServiceClient
	recordingService   ros2bridgev1.RecordingServiceClient
	sensingService     ros2bridgev1.SensingServiceClient
	timeout            time.Duration
}

func NewClient(address string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	return &Client{
		conn:               conn,
		serviceManager:     modulev1.NewServiceManagerServiceClient(conn),
		systemControl:      modulev1.NewSystemControlServiceClient(conn),
		monitoring:         modulev1.NewMonitoringServiceClient(conn),
		recordingService:   ros2bridgev1.NewRecordingServiceClient(conn),
		sensingService:     ros2bridgev1.NewSensingServiceClient(conn),
		timeout:            timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

// ServiceManagerService methods
func (c *Client) ListServices() (*modulev1.ListServicesResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.ListServicesRequest{}
	return c.serviceManager.ListServices(ctx, req)
}

func (c *Client) GetService(name string) (*modulev1.Service, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.GetServiceRequest{Name: name}
	return c.serviceManager.GetService(ctx, req)
}

func (c *Client) StartService(name string) (*modulev1.StartServiceResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.StartServiceRequest{Name: name}
	return c.serviceManager.StartService(ctx, req)
}

func (c *Client) StopService(name string) (*modulev1.StopServiceResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.StopServiceRequest{Name: name}
	return c.serviceManager.StopService(ctx, req)
}

func (c *Client) RestartService(name string) (*modulev1.RestartServiceResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.RestartServiceRequest{Name: name}
	return c.serviceManager.RestartService(ctx, req)
}

func (c *Client) EnableService(name string) (*modulev1.EnableServiceResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.EnableServiceRequest{Name: name}
	return c.serviceManager.EnableService(ctx, req)
}

func (c *Client) DisableService(name string) (*modulev1.DisableServiceResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.DisableServiceRequest{Name: name}
	return c.serviceManager.DisableService(ctx, req)
}

// SystemControlService methods
func (c *Client) Reboot(delaySeconds int32) (*modulev1.RebootResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.RebootRequest{DelaySeconds: delaySeconds}
	return c.systemControl.Reboot(ctx, req)
}

func (c *Client) Shutdown(delaySeconds int32) (*modulev1.ShutdownResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.ShutdownRequest{DelaySeconds: delaySeconds}
	return c.systemControl.Shutdown(ctx, req)
}

// MonitoringService methods
func (c *Client) GetDiskUsage() (*modulev1.GetDiskUsageResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.GetDiskUsageRequest{}
	return c.monitoring.GetDiskUsage(ctx, req)
}

func (c *Client) GetPTPStatus(includeRemote bool) (*modulev1.GetPTPStatusResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &modulev1.GetPTPStatusRequest{IncludeRemoteDevices: includeRemote}
	return c.monitoring.GetPTPStatus(ctx, req)
}

// SensingService methods
func (c *Client) GetPosition() (*ros2bridgev1.Position, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.GetPositionRequest{}
	return c.sensingService.GetPosition(ctx, req)
}

func (c *Client) ListNodes(filter string) (*ros2bridgev1.ListNodesResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.ListNodesRequest{Filter: filter}
	return c.sensingService.ListNodes(ctx, req)
}

// RecordingService methods
func (c *Client) StartRecording() (*ros2bridgev1.StartRecordingResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.StartRecordingRequest{}
	return c.recordingService.StartRecording(ctx, req)
}

func (c *Client) StopRecording() (*ros2bridgev1.StopRecordingResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.StopRecordingRequest{}
	return c.recordingService.StopRecording(ctx, req)
}

func (c *Client) PauseRecording() (*ros2bridgev1.PauseRecordingResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.PauseRecordingRequest{}
	return c.recordingService.PauseRecording(ctx, req)
}

func (c *Client) ResumeRecording() (*ros2bridgev1.ResumeRecordingResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.ResumeRecordingRequest{}
	return c.recordingService.ResumeRecording(ctx, req)
}

func (c *Client) GetRecording(hardwareID string) (*ros2bridgev1.Recording, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.GetRecordingRequest{HardwareId: hardwareID}
	return c.recordingService.GetRecording(ctx, req)
}

func (c *Client) ListRecordings(filter string) (*ros2bridgev1.ListRecordingsResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.ListRecordingsRequest{Filter: filter}
	return c.recordingService.ListRecordings(ctx, req)
}

func (c *Client) ListTopicStatuses(hardwareID, filter string) (*ros2bridgev1.ListTopicStatusesResponse, error) {
	ctx, cancel := c.GetContext()
	defer cancel()

	req := &ros2bridgev1.ListTopicStatusesRequest{
		HardwareId: hardwareID,
		Filter:     filter,
	}
	return c.recordingService.ListTopicStatuses(ctx, req)
}
//go:build !lite

package service

import (
	"context"
	"fmt"
	"log"

	systemv1 "github.com/proto_api/services/system-manager/gen/system/v1"
)

func (s *SystemService) ManageDrsService(ctx context.Context, req *systemv1.ManageDrsServiceRequest) (*systemv1.ManageDrsServiceResponse, error) {
	if !s.config.Services.EnableSystemdManage {
		return &systemv1.ManageDrsServiceResponse{
			Success: false,
			Message: "systemd service management is not enabled in this mode",
		}, nil
	}

	log.Printf("DRS service management request: %v", req.Action)
	
	var err error
	var status string
	
	switch req.Action {
	case systemv1.ManageDrsServiceRequest_SERVICE_ACTION_STOP:
		err = s.systemManager.StopDrsService()
		status = "stopped"
	case systemv1.ManageDrsServiceRequest_SERVICE_ACTION_RESTART:
		err = s.systemManager.RestartDrsService()
		status = "restarted"
	case systemv1.ManageDrsServiceRequest_SERVICE_ACTION_STATUS:
		status, err = s.systemManager.GetDrsServiceStatus()
	default:
		return &systemv1.ManageDrsServiceResponse{
			Success: false,
			Message: "invalid action specified",
		}, nil
	}

	if err != nil {
		return &systemv1.ManageDrsServiceResponse{
			Success: false,
			Message: fmt.Sprintf("DRS service operation failed: %v", err),
		}, nil
	}

	return &systemv1.ManageDrsServiceResponse{
		Success:       true,
		Message:       fmt.Sprintf("DRS service operation completed successfully"),
		ServiceStatus: status,
	}, nil
}

func (s *SystemService) ManageRecorderService(ctx context.Context, req *systemv1.ManageRecorderServiceRequest) (*systemv1.ManageRecorderServiceResponse, error) {
	if !s.config.Services.EnableSystemdManage {
		return &systemv1.ManageRecorderServiceResponse{
			Success: false,
			Message: "systemd service management is not enabled in this mode",
		}, nil
	}

	log.Printf("Recorder service management request: %v", req.Action)
	
	var err error
	var status string
	
	switch req.Action {
	case systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_STOP:
		err = s.systemManager.StopRecorderService()
		status = "stopped"
	case systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_RESTART:
		err = s.systemManager.RestartRecorderService()
		status = "restarted"
	case systemv1.ManageRecorderServiceRequest_SERVICE_ACTION_STATUS:
		status, err = s.systemManager.GetRecorderServiceStatus()
	default:
		return &systemv1.ManageRecorderServiceResponse{
			Success: false,
			Message: "invalid action specified",
		}, nil
	}

	if err != nil {
		return &systemv1.ManageRecorderServiceResponse{
			Success: false,
			Message: fmt.Sprintf("Recorder service operation failed: %v", err),
		}, nil
	}

	return &systemv1.ManageRecorderServiceResponse{
		Success:       true,
		Message:       fmt.Sprintf("Recorder service operation completed successfully"),
		ServiceStatus: status,
	}, nil
}
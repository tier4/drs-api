//go:build lite

package service

import (
	"context"

	systemv1 "github.com/proto_api/services/system-manager/gen/system/v1"
)

func (s *SystemService) ManageDrsService(ctx context.Context, req *systemv1.ManageDrsServiceRequest) (*systemv1.ManageDrsServiceResponse, error) {
	// Lite version returns not supported error
	return &systemv1.ManageDrsServiceResponse{
		Success: false,
		Message: "drs service management is not available in lite version",
	}, nil
}

func (s *SystemService) ManageRecorderService(ctx context.Context, req *systemv1.ManageRecorderServiceRequest) (*systemv1.ManageRecorderServiceResponse, error) {
	// Lite version returns not supported error
	return &systemv1.ManageRecorderServiceResponse{
		Success: false,
		Message: "recorder service management is not available in lite version",
	}, nil
}
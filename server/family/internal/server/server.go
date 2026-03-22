package server

import (
	"context"

	familyv1 "github.com/boykush/famoney/server/family/gen/go"
)

type Server struct {
	familyv1.UnimplementedFamilyServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) HealthCheck(ctx context.Context, req *familyv1.HealthCheckRequest) (*familyv1.HealthCheckResponse, error) {
	return &familyv1.HealthCheckResponse{
		Status: familyv1.HealthCheckResponse_SERVING,
	}, nil
}

package server

import (
	"context"

	expensev1 "github.com/boykush/famoney/server/expense/gen/go"
)

type Server struct {
	expensev1.UnimplementedExpenseServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) HealthCheck(ctx context.Context, req *expensev1.HealthCheckRequest) (*expensev1.HealthCheckResponse, error) {
	return &expensev1.HealthCheckResponse{
		Status: expensev1.HealthCheckResponse_SERVING,
	}, nil
}

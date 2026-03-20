package provider

import (
	"fmt"
	"log"
	"net"

	"github.com/samber/do/v2"

	expensev1 "github.com/boykush/famoney/server/expense/gen/go"
	"github.com/boykush/famoney/server/expense/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer wraps *grpc.Server to implement do.Shutdownable.
type GRPCServer struct {
	*grpc.Server
}

func (s *GRPCServer) Shutdown() error {
	s.GracefulStop()
	return nil
}

// ProvideServer creates a new expense gRPC service server.
func ProvideServer(i do.Injector) (*server.Server, error) {
	return server.NewServer(), nil
}

// ProvideGRPCServer creates and starts a gRPC server with the expense service registered.
func ProvideGRPCServer(i do.Injector) (*GRPCServer, error) {
	cfg := do.MustInvoke[Config](i)
	srv := do.MustInvoke[*server.Server](i)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	expensev1.RegisterExpenseServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	log.Printf("Starting Expense server on port %s", cfg.Port)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Expense server failed to serve: %v", err)
		}
	}()

	return &GRPCServer{Server: grpcServer}, nil
}

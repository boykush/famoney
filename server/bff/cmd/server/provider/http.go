package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/samber/do/v2"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	expensepb "github.com/boykush/famoney/server/bff/gen/go/expense"
	familypb "github.com/boykush/famoney/server/bff/gen/go/family"
)

// ProvideHTTPServer creates and starts an HTTP server with gRPC-Gateway handlers.
func ProvideHTTPServer(i do.Injector) (*http.Server, error) {
	cfg := do.MustInvoke[Config](i)
	oidcVerifier := do.MustInvoke[*OIDCVerifier](i)

	ctx := context.Background()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := expensepb.RegisterExpenseServiceHandlerFromEndpoint(ctx, mux, cfg.ExpenseServiceAddr, opts); err != nil {
		return nil, fmt.Errorf("failed to register expense service handler: %w", err)
	}

	if err := familypb.RegisterFamilyServiceHandlerFromEndpoint(ctx, mux, cfg.FamilyServiceAddr, opts); err != nil {
		return nil, fmt.Errorf("failed to register family service handler: %w", err)
	}

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           oidcVerifier.Middleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting BFF server on port %s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return httpServer, nil
}

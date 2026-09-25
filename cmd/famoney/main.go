// famoney は家計データのパイプラインと、それを配る MCP サーバーをまとめた CLI。
// 1つのイメージをサブコマンドで使い分ける。
//
//	famoney ingest    [--month YYYY-MM|current|previous]  マネーフォワード ME の CSV を raw 層へ
//	famoney transform [--month YYYY-MM|current|previous]  raw 層の CSV を product の明細へ
//	famoney mcp       [--addr host:port]                  明細を MCP で配る
//	famoney duckdb-extensions <dir>                       DuckDB の拡張を dir に入れる（イメージのビルド用）
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/boykush/famoney/internal/ingest"
	"github.com/boykush/famoney/internal/lake"
	"github.com/boykush/famoney/internal/mcpserver"
	"github.com/boykush/famoney/internal/moneyforward"
	"github.com/boykush/famoney/internal/month"
	"github.com/boykush/famoney/internal/transform"
)

// version はビルド時に ldflags で埋める。
var version = "dev"

const usage = `usage: famoney <command> [flags]

commands:
  ingest              download the Money Forward ME CSV of a month into the raw layer
  transform           convert the raw CSV of a month into the product transactions
  mcp                 serve the product transactions over MCP
  duckdb-extensions   install DuckDB extensions into a directory (image build)
  version             print the version
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:]); err != nil {
		slog.Error("famoney failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return errors.New("no command")
	}
	cmd, args := args[0], args[1:]

	switch cmd {
	case "ingest":
		return runIngest(ctx, args)
	case "transform":
		return runTransform(ctx, args)
	case "mcp":
		return runMCP(ctx, args)
	case "duckdb-extensions":
		if len(args) != 1 {
			return errors.New("usage: famoney duckdb-extensions <dir>")
		}
		return lake.InstallExtensions(ctx, args[0])
	case "version":
		fmt.Println(version)
		return nil
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func monthFlag(name string, args []string) (month.Month, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	m := fs.String("month", "current", "target month: YYYY-MM, current or previous (in JST)")
	if err := fs.Parse(args); err != nil {
		return month.Month{}, err
	}
	return month.Parse(*m, time.Now())
}

func runIngest(ctx context.Context, args []string) error {
	m, err := monthFlag("ingest", args)
	if err != nil {
		return err
	}
	l, err := openLake(ctx)
	if err != nil {
		return err
	}
	defer l.Close()

	client := moneyforward.NewClient(os.Getenv("MONEYFORWARD_COOKIE"))
	dst, err := ingest.Run(ctx, l, client, m)
	if err != nil {
		return err
	}
	slog.Info("ingested", "month", m.String(), "path", dst)
	return nil
}

func runTransform(ctx context.Context, args []string) error {
	m, err := monthFlag("transform", args)
	if err != nil {
		return err
	}
	l, err := openLake(ctx)
	if err != nil {
		return err
	}
	defer l.Close()

	n, dst, err := transform.Run(ctx, l, m)
	if err != nil {
		return err
	}
	slog.Info("transformed", "month", m.String(), "rows", n, "path", dst)
	return nil
}

func runMCP(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	addr := fs.String("addr", "0.0.0.0:8080", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tokens := splitList(os.Getenv("FAMONEY_MCP_TOKENS"))
	if len(tokens) == 0 {
		return errors.New("FAMONEY_MCP_TOKENS is empty: the MCP server never runs without authentication")
	}

	l, err := openLake(ctx)
	if err != nil {
		return err
	}
	defer l.Close()

	server := mcpserver.NewServer(mcpserver.NewStore(l), version)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           mcpserver.Handler(server, tokens),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	slog.Info("serving MCP", "addr", *addr, "version", version)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func openLake(ctx context.Context) (*lake.Lake, error) {
	return lake.Open(ctx, lake.Config{
		Root: os.Getenv("FAMONEY_LAKE_ROOT"),
		S3: lake.S3Config{
			Endpoint:        os.Getenv("FAMONEY_S3_ENDPOINT"),
			Region:          os.Getenv("FAMONEY_S3_REGION"),
			AccessKeyID:     os.Getenv("FAMONEY_S3_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("FAMONEY_S3_SECRET_ACCESS_KEY"),
		},
		ExtensionDirectory: os.Getenv("FAMONEY_DUCKDB_EXTENSION_DIRECTORY"),
	})
}

func splitList(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

package main

import (
	"context"
	"example.com/feature-rollout-control/internal/config"
	"example.com/feature-rollout-control/internal/server"
	d "example.com/feature-rollout-control/internal/shared/domain"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
	repo := d.NewMemoryRepository()
	pub := d.NewMemoryPublisher()
	svc := server.NewService(repo, d.SystemClock{}, d.StableHasher{}, pub, log)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg := config.Load()
	if err := server.Start(ctx, cfg.Address, svc, log); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

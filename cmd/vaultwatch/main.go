// Package main is the entry point for the vaultwatch CLI tool.
// It wires together configuration, Vault client, monitor, alerting, and scheduling.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/config"
	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/scheduler"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(logger); err != nil {
		logger.Error("vaultwatch exited with error", "error", err)
		os.Exit(1)
	}
}

// run contains the core application logic, separated from main for testability.
func run(logger *slog.Logger) error {
	// Determine config file path from env or use default.
	cfgPath := os.Getenv("VAULTWATCH_CONFIG")
	if cfgPath == "" {
		cfgPath = "vaultwatch.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger.Info("configuration loaded",
		"vault_addr", cfg.VaultAddr,
		"secret_count", len(cfg.SecretPaths),
		"check_interval", cfg.CheckInterval,
		"warn_threshold", cfg.WarnThreshold,
	)

	vaultClient, err := vault.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	// Perform an initial health check before starting the schedule loop.
	if err := vaultClient.IsHealthy(); err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	logger.Info("vault health check passed")

	sender := alert.NewStdoutSender(os.Stdout)

	mon := monitor.New(cfg, vaultClient, sender, logger)

	sched := scheduler.New(cfg.CheckInterval, mon, logger)

	// Set up context that cancels on SIGINT or SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("starting vaultwatch", "interval", cfg.CheckInterval)

	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("scheduler error: %w", err)
	}

	logger.Info("vaultwatch stopped gracefully")
	return nil
}

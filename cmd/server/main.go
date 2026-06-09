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
	"syscall"
	"time"

	"github.com/s3ntin3l8/go-http-template/internal/config"
	"github.com/s3ntin3l8/go-http-template/internal/httpapi"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	addr := flag.String("addr", "", "override listen address (default from config)")
	healthcheck := flag.Bool("healthcheck", false, "run health check and exit")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if *healthcheck {
		listenAddr := cfg.ListenAddr
		if *addr != "" {
			listenAddr = *addr
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://" + listenAddr + "/health")
		if err != nil {
			fmt.Fprintf(os.Stderr, "health check failed: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "health check returned status %d\n", resp.StatusCode)
			os.Exit(1)
		}
		fmt.Println("healthy")
		return
	}

	listenAddr := cfg.ListenAddr
	if *addr != "" {
		listenAddr = *addr
	}
	cfg.ListenAddr = listenAddr

	srv := httpapi.New(cfg)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("starting server", "addr", listenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
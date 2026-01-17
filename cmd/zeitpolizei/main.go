package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nadilas/zeitpolizei/internal/api"
	"github.com/nadilas/zeitpolizei/internal/config"
	"github.com/nadilas/zeitpolizei/internal/enforcer"
	"github.com/nadilas/zeitpolizei/internal/storage"
	"github.com/nadilas/zeitpolizei/internal/tracker"
	"github.com/nadilas/zeitpolizei/internal/unifi"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("zeitpolizei %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.NewSQLite(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize UniFi client
	unifiClient, err := unifi.NewClient(unifi.Config{
		BaseURL:  cfg.UniFi.URL,
		Username: cfg.UniFi.Username,
		Password: cfg.UniFi.Password,
		Site:     cfg.UniFi.Site,
		IsUDM:    cfg.UniFi.IsUDM,
		Insecure: cfg.UniFi.Insecure,
	})
	if err != nil {
		log.Fatalf("Failed to initialize UniFi client: %v", err)
	}

	// Login to UniFi controller
	if err := unifiClient.Login(); err != nil {
		log.Fatalf("Failed to login to UniFi controller: %v", err)
	}
	log.Println("Successfully connected to UniFi controller")

	// Initialize enforcer
	enf := enforcer.New(store, unifiClient)

	// Initialize tracker
	track := tracker.New(store, unifiClient, enf, cfg.Tracker.PollInterval)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start tracker
	go track.Start(ctx)

	// Initialize and start API server
	server := api.NewServer(cfg, store, unifiClient, enf)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		server.Shutdown()
	}()

	// Start HTTP server (blocking)
	log.Printf("Starting Zeitpolizei on %s", cfg.Server.Address)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

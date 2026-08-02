package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/saitama-op/healthcheck-agent/internal/checker"
	"github.com/saitama-op/healthcheck-agent/internal/config"
)

func main() {
	configPath := flag.String("config", "configs/healthcheck.yaml", "Path to configuration file")
	bindInterface := flag.String("interface", "", "Override network interface to bind to (e.g., eth0)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		if *verbose {
			fmt.Printf("Configuration Error: %v\n", err)
		}
		os.Exit(2)
	}

	// CLI flag overrides YAML config
	if *bindInterface != "" {
		cfg.BindInterface = *bindInterface
	}

	// Create a master context that strictly enforces the global timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	exitCode := checker.Run(ctx, cfg, *verbose)
	
	if !*verbose {
		fmt.Printf("Exit %d\n", exitCode)
	} else {
		fmt.Printf("Exit Code: %d\n", exitCode)
	}
	
	os.Exit(exitCode)
}

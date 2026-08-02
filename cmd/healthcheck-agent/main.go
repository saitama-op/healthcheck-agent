package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/healthcheck-agent/internal/checker"
	"github.com/healthcheck-agent/internal/config"
)

func main() {
	configPath := flag.String("config", "configs/healthcheck.yaml", "Path to configuration file")
	bindInterface := flag.String("interface", "", "Override network interface to bind to (e.g., eth0)")
	dnsResolver := flag.String("dns", "", "Override DNS resolver to use (e.g., 8.8.8.8)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		if *verbose {
			fmt.Printf("Configuration Error: %v\n", err)
		}
		os.Exit(2)
	}

	// CLI flags override YAML config
	if *bindInterface != "" {
		cfg.BindInterface = *bindInterface
	}
	if *dnsResolver != "" {
		cfg.DNSResolver = *dnsResolver
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

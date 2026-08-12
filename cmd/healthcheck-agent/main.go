package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/saitama-op/healthcheck-agent/internal/checker"
	"github.com/saitama-op/healthcheck-agent/internal/config"
	"os"
	"time"
)

// resolveConfigPath checks multiple standard locations for the config file
func resolveConfigPath(flagPath string) string {
	// 1. If the user explicitly provided a path via the flag, always use it
	if flagPath != "" {
		return flagPath
	}

	// 2. Define standard fallback paths
	searchPaths := []string{
		"configs/healthcheck.yaml",                     // Local development path
		"/etc/healthcheck-agent/healthcheck.yaml",      // Standard Linux config path
		"/usr/local/etc/healthcheck-agent/config.yaml", // Alternate Linux config path
	}

	// 3. Return the first file that actually exists
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 4. Default fallback (will throw a clean error in config.Load if it doesn't exist)
	return "configs/healthcheck.yaml"
}

func main() {
	// Default is now empty so we can detect if the user explicitly set it
	configFlag := flag.String("config", "", "Path to configuration file (default searches ./configs/ and /etc/)")
	bindInterface := flag.String("interface", "", "Override network interface to bind to (e.g., eth0)")
	dnsResolver := flag.String("dns", "", "Override DNS resolver to use (e.g., 8.8.8.8)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Automatically find the best config path
	configPath := resolveConfigPath(*configFlag)

	cfg, err := config.Load(configPath)
	if err != nil {
		if *verbose {
			fmt.Printf("Configuration Error: %v (Checked path: %s)\n", err, configPath)
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
	totalAttempts := time.Duration(cfg.Retry + 1)
	maxRunTime := (totalAttempts * cfg.Timeout) + (time.Duration(cfg.Retry) * cfg.RetryDelay) + (2 * time.Second)

	// Create a master context using the total maxRunTime
	ctx, cancel := context.WithTimeout(context.Background(), maxRunTime)
	defer cancel()

	exitCode := checker.Run(ctx, cfg, *verbose)

	if !*verbose {
		fmt.Printf("Exit %d\n", exitCode)
	} else {
		fmt.Printf("Exit Code: %d\n", exitCode)
	}

	os.Exit(exitCode)
}

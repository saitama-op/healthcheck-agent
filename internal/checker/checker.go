package checker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/saitama-op/healthcheck-agent/internal/config"
)

type Result struct {
	URL        string
	Expected   int
	Actual     int
	Passed     bool
	Error      error
}

// getLocalIP finds the first valid IPv4 address for a given interface name
func getLocalIP(ifaceName string) (net.IP, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("interface %s not found: %w", ifaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil || len(addrs) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", ifaceName)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return nil, fmt.Errorf("no valid IPv4 address found on %s", ifaceName)
}

// Run executes the health checks concurrently
func Run(ctx context.Context, cfg *config.Config, verbose bool) int {
	dialer := &net.Dialer{
		Timeout: cfg.Timeout,
	}

	var localIP net.IP
	var err error

	// Extract IP if binding to a specific interface
	if cfg.BindInterface != "" {
		localIP, err = getLocalIP(cfg.BindInterface)
		if err != nil {
			if verbose {
				fmt.Printf("Internal Error: %v\n", err)
			}
			return 3 // Internal Error
		}
		// Bind standard HTTP (TCP) traffic to the interface
		dialer.LocalAddr = &net.TCPAddr{IP: localIP}
	}

	// Set up Custom DNS Resolver if configured
	if cfg.DNSResolver != "" {
		dnsServer := cfg.DNSResolver
		// Automatically append default DNS port if not provided
		if !strings.Contains(dnsServer, ":") {
			dnsServer = net.JoinHostPort(dnsServer, "53")
		}

		dialer.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dnsDialer := net.Dialer{
					Timeout: cfg.Timeout,
				}
				// Force DNS (UDP) traffic out the specific interface too
				if localIP != nil {
					dnsDialer.LocalAddr = &net.UDPAddr{IP: localIP}
				}
				return dnsDialer.DialContext(ctx, "udp", dnsServer)
			},
		}
	}

	transport := &http.Transport{
		DialContext:       dialer.DialContext,
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	results := make(chan Result, len(cfg.URLs))
	var wg sync.WaitGroup

	// Spawn concurrent checks
	for _, u := range cfg.URLs {
		wg.Add(1)
		go func(urlConfig config.URLConfig) {
			defer wg.Done()
			results <- checkURL(ctx, client, urlConfig, cfg.UserAgent)
		}(u)
	}

	wg.Wait()
	close(results)

	passed := 0
	total := len(cfg.URLs)

	if verbose {
		fmt.Printf("Checking (Interface: %s, DNS: %s):\n", cfg.BindInterface, cfg.DNSResolver)
	}

	for res := range results {
		if res.Passed {
			passed++
			if verbose {
				fmt.Printf("✓ %s (%d)\n", res.URL, res.Actual)
			}
		} else {
			if verbose {
				errStr := ""
				if res.Error != nil {
					errStr = fmt.Sprintf(" - %v", res.Error)
				}
				fmt.Printf("✗ %s (Expected %d, Got %d)%s\n", res.URL, res.Expected, res.Actual, errStr)
			}
		}
	}

	successPercent := (float64(passed) / float64(total)) * 100

	if verbose {
		fmt.Printf("\nSuccess: %d/%d (%.0f%%)\n", passed, total, successPercent)
	}

	if successPercent >= cfg.MinimumSuccessPercent {
		return 0 // Healthy
	}
	return 1 // Unhealthy
}

func checkURL(ctx context.Context, client *http.Client, cfg config.URLConfig, userAgent string) Result {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return Result{URL: cfg.URL, Expected: cfg.ExpectedStatus, Passed: false, Error: err}
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: cfg.URL, Expected: cfg.ExpectedStatus, Passed: false, Error: err}
	}
	defer resp.Body.Close()

	passed := resp.StatusCode == cfg.ExpectedStatus
	return Result{
		URL:      cfg.URL,
		Expected: cfg.ExpectedStatus,
		Actual:   resp.StatusCode,
		Passed:   passed,
	}
}

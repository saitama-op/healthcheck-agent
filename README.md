# healthcheck-agent

A lightweight, dependency-free Go binary designed to validate true internet connectivity across multi-WAN and complex routing environments such as VyOS, pfSense, and standard Linux.

Traditional WAN health checks often rely on pinging a single endpoint such as `8.8.8.8`. This can cause false-positive WAN failovers when a single datacenter is unavailable, ICMP is deprioritized, or a specific destination temporarily has connectivity problems.

`healthcheck-agent` solves this by polling multiple industry-standard connectivity and captive-portal endpoints from different providers and calculating an overall success percentage.

---

## Features

- **Lightweight & Fast:** Single statically linked Go binary.
- **Multi-WAN Native:** Force HTTP/HTTPS and DNS traffic through a specific physical interface using Linux `SO_BINDTODEVICE`.
- **Multiple Independent Endpoints:** Test connectivity against Cloudflare, Apple, Microsoft, Mozilla, Google, and other providers.
- **Resiliency Threshold:** Configure `minimum_success_percent` to prevent unnecessary failovers when one or more test endpoints are temporarily unavailable.
- **Custom DNS Resolution:** Specify a DNS resolver per WAN interface to validate both DNS and Internet connectivity.
- **Retry Logic:** Built-in retries and configurable retry delays help tolerate transient packet loss and short-lived network issues.
- **Expected HTTP Status:** Each endpoint can define the HTTP status code expected for a successful connectivity test.
- **Deterministic Exit Codes:** Returns `0` for healthy and `1` for unhealthy, making it suitable for shell scripts, cron, Keepalived, and VyOS WAN load balancing.
- **Verbose Diagnostics:** Optional verbose mode displays individual endpoint results and overall health statistics.
- **No External Runtime Dependencies:** The application is compiled as a standalone Go binary.

---

## How It Works

For each configured endpoint, the agent:

1. Resolves the hostname using the configured DNS resolver.
2. Binds network traffic to the specified physical interface when configured.
3. Connects to the endpoint using HTTP or HTTPS.
4. Validates the returned HTTP status code.
5. Retries failed checks according to the configured retry policy.
6. Calculates the overall success percentage.
7. Returns a deterministic exit code.

For example, with:

```yaml
minimum_success_percent: 60.0
```

and 11 configured endpoints:

```text
11 endpoints
    |
    +-- Endpoint 1  -> success
    +-- Endpoint 2  -> success
    +-- Endpoint 3  -> success
    +-- Endpoint 4  -> success
    +-- Endpoint 5  -> success
    +-- Endpoint 6  -> success
    +-- Endpoint 7  -> success
    +-- Endpoint 8  -> failure
    +-- Endpoint 9  -> failure
    +-- Endpoint 10 -> failure
    +-- Endpoint 11 -> failure
    |
    +-- 7/11 successful
    |
    +-- 63.6%
    |
    +-- HEALTHY
```

This prevents a single endpoint or provider outage from unnecessarily triggering WAN failover.

---

## Installation

### From Source

Install Go from the official documentation:

https://go.dev/doc/install

Clone the repository:

```bash
git clone https://github.com/saitama-op/healthcheck-agent.git
cd healthcheck-agent
```

Build the binary:

```bash
make
```

The resulting binary can then be copied to the desired location:

```bash
sudo cp healthcheck-agent /usr/local/bin/healthcheck-agent
sudo chmod +x /usr/local/bin/healthcheck-agent
```

---

## Release Builds

To compile release binaries for multiple operating systems and architectures:

```bash
make release
```

The release process can produce binaries for platforms such as:

- Linux AMD64
- Linux ARM64
- macOS AMD64
- macOS ARM64
- Windows AMD64
- Windows ARM64

Refer to the Makefile for the exact supported build targets.

---

## Configuration

The agent accepts a configuration file using the `-config` option.

Example:

```bash
./healthcheck-agent -config /etc/healthcheck-agent/healthcheck.yaml
```

The default configuration lookup behavior is:

```text
./configs/
 /etc/
```

A custom configuration file can always be supplied using:

```bash
-config /path/to/healthcheck.yaml
```

---

## Configuration File

Example `healthcheck.yaml`:

```yaml
# Global settings

# Maximum time allowed for an individual HTTP/HTTPS request.
timeout: 250ms

# Minimum percentage of endpoints that must succeed
# for the WAN to be considered healthy.
minimum_success_percent: 60.0

# User-Agent sent with HTTP/HTTPS requests.
user_agent: "Dalvik/2.1.0 (Linux; U; Android 14; Pixel 8 Pro)"

# Optional interface binding.
# Forces connectivity tests through a specific WAN interface.
bind_interface: eth0

# Optional DNS resolver.
# Used to resolve endpoint hostnames through the ISP's DNS resolver.
dns_resolver: 192.168.0.1

# Number of retries for failed checks.
retry: 3

# Delay between retries.
retry_delay: 250ms

# Connectivity endpoints.
urls:
  - url: http://cp.cloudflare.com/generate_204
    expected_status: 204

  - url: http://detectportal.firefox.com/success.txt
    expected_status: 200

  - url: http://captive.apple.com/hotspot-detect.html
    expected_status: 200

  - url: http://www.msftconnecttest.com/connecttest.txt
    expected_status: 200

  - url: https://www.gstatic.com/generate_204
    expected_status: 204

  - url: https://clients3.google.com/generate_204
    expected_status: 204

  - url: https://connectivitycheck.gstatic.com/generate_204
    expected_status: 204

  - url: https://www.google.com/generate_204
    expected_status: 204

  - url: https://www.apple.com
    expected_status: 200

  - url: https://1.1.1.1
    expected_status: 200

  - url: https://www.microsoft.com
    expected_status: 200
```

---

## Configuration Options

| Option | Description |
|---|---|
| `timeout` | Maximum duration allowed for an individual HTTP/HTTPS request. |
| `minimum_success_percent` | Minimum percentage of endpoints that must succeed for the overall test to be healthy. |
| `user_agent` | HTTP User-Agent header sent to endpoints. |
| `bind_interface` | Optional Linux interface to which network traffic is bound. |
| `dns_resolver` | Optional DNS resolver used for hostname resolution. |
| `retry` | Number of retries performed for failed endpoint checks. |
| `retry_delay` | Delay between retry attempts. |
| `urls` | List of endpoints to test. |
| `urls[].url` | Endpoint URL. |
| `urls[].expected_status` | HTTP status code expected for a successful request. |

---

## Multi-WAN Interface Binding

The `bind_interface` option allows the health check to be executed through a specific WAN interface.

For example:

```yaml
bind_interface: eth3
dns_resolver: 192.168.1.1
```

This ensures that the health check validates the Internet connection associated with `eth3` rather than relying on the system's default route.

This is particularly useful on multi-WAN routers such as VyOS.

Example:

```text
                    +-- eth0 -- ISP 1
                    |
LAN ---- VyOS ------+-- eth2 -- ISP 2
                    |
                    +-- eth3 -- ISP 3
```

Each WAN can use its own configuration:

```text
healthcheck-act.yaml
    bind_interface: eth0
    dns_resolver: 192.168.0.1

healthcheck-hathway.yaml
    bind_interface: eth2
    dns_resolver: 192.168.5.1

healthcheck-airtel.yaml
    bind_interface: eth3
    dns_resolver: 192.168.1.1
```

---

## Usage

### Standard / Quiet Mode

Ideal for automation and shell scripts:

```bash
./healthcheck-agent
```

The program returns an exit code without requiring verbose output.

Check the result:

```bash
echo $?
```

A result of `0` means the WAN is healthy.

A result of `1` means the WAN is unhealthy.

---

### Verbose Mode

Use verbose mode when troubleshooting:

```bash
./healthcheck-agent -verbose
```

Example output:

```text
Checking (Interface: eth0, DNS: 192.168.0.1):
✓ http://www.msftconnecttest.com/connecttest.txt (200)
✓ http://cp.cloudflare.com/generate_204 (204)
✓ http://detectportal.firefox.com/success.txt (200)
✓ http://captive.apple.com/hotspot-detect.html (200)
✓ https://www.google.com/generate_204 (204)
✓ https://www.apple.com (200)
✓ https://www.gstatic.com/generate_204 (204)
✓ https://connectivitycheck.gstatic.com/generate_204 (204)
✓ https://clients3.google.com/generate_204 (204)
✓ https://1.1.1.1 (200)
✓ https://www.microsoft.com (200)

Success: 11/11 (100%)
Exit Code: 0
```

---

## Override Interface and DNS

The interface and DNS resolver can be overridden from the command line:

```bash
sudo ./healthcheck-agent \
  -verbose \
  -interface eth2 \
  -dns 1.1.1.1
```

This is useful for testing individual WAN connections without changing the configuration file.

---

## Linux Interface Binding

When `bind_interface` or the `-interface` option is used, the application uses Linux `SO_BINDTODEVICE` to bind network traffic to the specified interface.

For example:

```bash
-interface eth2
```

forces the connectivity tests through `eth2`.

Binding to a physical interface may require elevated privileges.

Run with:

```bash
sudo ./healthcheck-agent -interface eth2
```

or grant the appropriate capability to the binary if required by the deployment environment.

---

## Exit Codes

The agent is designed to be consumed by automation systems.

| Exit Code | Meaning | Description |
|---:|---|---|
| `0` | Healthy | The calculated success percentage is greater than or equal to `minimum_success_percent`. |
| `1` | Unhealthy | The calculated success percentage is below `minimum_success_percent`. |
| `2` | Configuration Error | Configuration file is missing, invalid, or contains invalid/empty endpoint configuration. |
| `3` | Internal Error | An internal error occurred, such as an unavailable network interface or socket binding failure. |

---

## Real-World Example: VyOS WAN Failover

The agent can be integrated directly with VyOS WAN load balancing using a small shell wrapper for each ISP.

### Airtel Example

Create:

```text
/config/scripts/healthcheck-airtel.sh
```

Example:

```bash
#!/bin/bash

/opt/healthcheck-agent/healthcheck-agent \
  -config /opt/healthcheck-agent/configs/healthcheck-airtel.yaml \
  -interface eth3 \
  -dns 192.168.1.1
```

Make it executable:

```bash
chmod +x /config/scripts/healthcheck-airtel.sh
```

### Hathway Example

```bash
#!/bin/bash

/opt/healthcheck-agent/healthcheck-agent \
  -config /opt/healthcheck-agent/configs/healthcheck-hathway.yaml \
  -interface eth2 \
  -dns 192.168.5.1
```

### ACT Example

```bash
#!/bin/bash

/opt/healthcheck-agent/healthcheck-agent \
  -config /opt/healthcheck-agent/configs/healthcheck-act.yaml \
  -interface eth0 \
  -dns 192.168.0.1
```

---

## VyOS Integration

Example VyOS WAN load-balancing configuration:

```text
set load-balancing wan interface-health eth0 failure-count '1'
set load-balancing wan interface-health eth0 nexthop 'dhcp'
set load-balancing wan interface-health eth0 success-count '5'
set load-balancing wan interface-health eth0 test 1 test-script '/config/scripts/healthcheck-act.sh'
set load-balancing wan interface-health eth0 test 1 type 'user-defined'

set load-balancing wan interface-health eth2 failure-count '1'
set load-balancing wan interface-health eth2 nexthop 'dhcp'
set load-balancing wan interface-health eth2 success-count '5'
set load-balancing wan interface-health eth2 test 1 test-script '/config/scripts/healthcheck-hathway.sh'
set load-balancing wan interface-health eth2 test 1 type 'user-defined'

set load-balancing wan interface-health eth3 failure-count '1'
set load-balancing wan interface-health eth3 nexthop 'dhcp'
set load-balancing wan interface-health eth3 success-count '5'
set load-balancing wan interface-health eth3 test 1 test-script '/config/scripts/healthcheck-airtel.sh'
set load-balancing wan interface-health eth3 test 1 type 'user-defined'
```

With this configuration:

```text
failure-count = 1
success-count = 5
```

the intended behavior is:

```text
WAN failure
    |
    v
First failed health-check result
    |
    v
WAN marked unhealthy
    |
    v
Failover to another WAN
```

When the WAN recovers:

```text
WAN recovery
    |
    v
Successful health check #1
    |
    v
Successful health check #2
    |
    v
Successful health check #3
    |
    v
Successful health check #4
    |
    v
Successful health check #5
    |
    v
WAN considered healthy
    |
    v
Failback becomes possible
```

This provides fast failover while requiring multiple consecutive successful health checks before the WAN is considered stable again.

---

## Recommended Multi-WAN Health Check Strategy

For a multi-WAN environment, avoid relying on a single endpoint such as:

```text
8.8.8.8
```

Instead, use multiple endpoints from different providers.

A good endpoint set should include:

- Cloudflare
- Google
- Microsoft
- Apple
- Mozilla
- Other independent connectivity endpoints

Prefer lightweight endpoints that return small responses such as:

```text
HTTP 204
HTTP 200
```

rather than large web pages.

This reduces bandwidth consumption and keeps health-check execution fast.

---

## Example Health Check Policy

A practical multi-WAN configuration could use:

```yaml
timeout: 250ms
minimum_success_percent: 60.0
retry: 3
retry_delay: 250ms
```

with approximately 10–12 independent endpoints.

For example, with 11 endpoints:

```text
11 endpoints
60% threshold
```

At least 7 successful endpoint checks are required to consider the WAN healthy.

Examples:

```text
11/11 → Healthy
10/11 → Healthy
 9/11 → Healthy
 8/11 → Healthy
 7/11 → Healthy
 6/11 → Unhealthy
```

This prevents one or two individual endpoint failures from unnecessarily causing WAN failover.

---

## Performance

The healthcheck-agent is designed to be lightweight and suitable for frequent WAN monitoring.

Example execution times on a VyOS multi-WAN router:

```text
Hathway:
11/11 successful
~223 ms

Airtel:
11/11 successful
~234 ms

ACT:
11/11 successful
~146 ms
```

The exact execution time depends on the ISP, DNS resolver, endpoint latency, TLS negotiation, and network conditions.

---

## Troubleshooting

### Check the Configuration

```bash
cat /etc/healthcheck-agent/healthcheck.yaml
```

Or specify a configuration explicitly:

```bash
./healthcheck-agent \
  -config /path/to/healthcheck.yaml \
  -verbose
```

### Test a Specific WAN Interface

```bash
sudo ./healthcheck-agent \
  -verbose \
  -interface eth2 \
  -dns 192.168.5.1
```

### Check the Exit Code

```bash
./healthcheck-agent
echo $?
```

Expected results:

```text
0 = Healthy
1 = Unhealthy
2 = Configuration Error
3 = Internal Error
```

### Test DNS Independently

```bash
nslookup www.google.com 192.168.5.1
```

or:

```bash
dig @192.168.5.1 www.google.com
```

### Check the WAN Interface

```bash
ip addr show eth2
```

### Check Routing

```bash
ip route
```

---

## Design Philosophy

The goal of `healthcheck-agent` is not simply to answer:

> "Can I ping the Internet?"

Instead, it tries to answer:

> "Does this specific WAN interface currently have usable Internet connectivity?"

That distinction is important in multi-WAN environments.

A WAN can have:

- A working physical link
- A working DHCP lease
- A reachable gateway
- Successful ICMP to the gateway

while still having no usable Internet connectivity.

By testing multiple external HTTP/HTTPS endpoints through the specific WAN interface, the agent provides a more realistic measurement of Internet availability.

---

## Security Considerations

The agent does not require inbound network access.

It performs outbound connectivity checks only.

When using interface binding:

```text
SO_BINDTODEVICE
```

the process may require elevated Linux privileges.

Only grant the minimum privileges required for the deployment environment.

---

## License

This project is licensed under the MIT License.

Copyright (c) 2026 sanjay kamalakshan

See the full license text in [LICENSE](LICENSE).

---

## Repository

GitHub:

https://github.com/saitama-op/healthcheck-agent


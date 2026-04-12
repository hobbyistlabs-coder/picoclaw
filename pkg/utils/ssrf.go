package utils

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
)

// allowPrivateWebFetchHosts controls whether loopback/private hosts are allowed.
// This is false in normal runtime to reduce SSRF exposure, and tests can override it temporarily.
var allowPrivateWebFetchHosts atomic.Bool

// AllowPrivateWebFetchHosts sets the toggle for whether private hosts are permitted in dial contexts.
func AllowPrivateWebFetchHosts(allow bool) {
	allowPrivateWebFetchHosts.Store(allow)
}

// GetAllowPrivateWebFetchHosts returns the current toggle state.
func GetAllowPrivateWebFetchHosts() bool {
	return allowPrivateWebFetchHosts.Load()
}

// NewSafeDialContext re-resolves DNS at connect time to mitigate DNS rebinding (TOCTOU)
// where a hostname resolves to a public IP during pre-flight but a private IP at connect time.
func NewSafeDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if allowPrivateWebFetchHosts.Load() {
			return dialer.DialContext(ctx, network, address)
		}

		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("invalid target address %q: %w", address, err)
		}
		if host == "" {
			return nil, fmt.Errorf("empty target host")
		}

		if ip := net.ParseIP(host); ip != nil {
			if IsPrivateOrRestrictedIP(ip) {
				return nil, fmt.Errorf("blocked private or local target: %s", host)
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		}

		ipAddrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", host, err)
		}

		attempted := 0
		var lastErr error
		for _, ipAddr := range ipAddrs {
			if IsPrivateOrRestrictedIP(ipAddr.IP) {
				continue
			}
			attempted++
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ipAddr.IP.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}

		if attempted == 0 {
			return nil, fmt.Errorf("all resolved addresses for %s are private or restricted", host)
		}
		if lastErr != nil {
			return nil, fmt.Errorf("failed connecting to public addresses for %s: %w", host, lastErr)
		}
		return nil, fmt.Errorf("failed connecting to public addresses for %s", host)
	}
}

// IsObviousPrivateHost performs a lightweight, no-DNS check for obviously private hosts.
// It catches localhost, literal private IPs, and empty hosts. It does NOT resolve DNS —
// the real SSRF guard is NewSafeDialContext which checks IPs at connect time.
func IsObviousPrivateHost(host string) bool {
	if allowPrivateWebFetchHosts.Load() {
		return false
	}

	h := strings.ToLower(strings.TrimSpace(host))
	h = strings.TrimSuffix(h, ".")
	if h == "" {
		return true
	}

	if h == "localhost" || strings.HasSuffix(h, ".localhost") {
		return true
	}

	if ip := net.ParseIP(h); ip != nil {
		return IsPrivateOrRestrictedIP(ip)
	}

	return false
}

// IsPrivateOrRestrictedIP returns true for IPs that should never be reached:
// RFC 1918, loopback, link-local (incl. cloud metadata 169.254.x.x), carrier-grade NAT,
// IPv6 unique-local (fc00::/7), 6to4 (2002::/16), and Teredo (2001:0000::/32).
func IsPrivateOrRestrictedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		// IPv4 private, loopback, link-local, and carrier-grade NAT ranges.
		if ip4[0] == 10 ||
			ip4[0] == 127 ||
			ip4[0] == 0 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168) ||
			(ip4[0] == 169 && ip4[1] == 254) ||
			(ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127) {
			return true
		}
		return false
	}

	if len(ip) == net.IPv6len {
		// IPv6 unique local addresses (fc00::/7)
		if (ip[0] & 0xfe) == 0xfc {
			return true
		}
		// 6to4 addresses (2002::/16): check the embedded IPv4 at bytes [2:6].
		if ip[0] == 0x20 && ip[1] == 0x02 {
			embedded := net.IPv4(ip[2], ip[3], ip[4], ip[5])
			return IsPrivateOrRestrictedIP(embedded)
		}
		// Teredo (2001:0000::/32): client IPv4 is at bytes [12:16], XOR-inverted.
		if ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x00 && ip[3] == 0x00 {
			client := net.IPv4(ip[12]^0xff, ip[13]^0xff, ip[14]^0xff, ip[15]^0xff)
			return IsPrivateOrRestrictedIP(client)
		}
	}

	return false
}

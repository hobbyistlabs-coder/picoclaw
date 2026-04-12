package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const maxRetries = 3

var retryDelayUnit = time.Second

func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode >= 500
}

func DoRequestWithRetry(client *http.Client, req *http.Request) (*http.Response, error) {
	// Ensure the client uses our SSRF-safe dialer if it's using the default transport or a custom transport without one
	safeClient := client
	if transport, ok := client.Transport.(*http.Transport); ok {
		if transport.DialContext == nil {
			// Shallow copy the transport to avoid modifying shared state
			safeTransport := transport.Clone()
			safeTransport.DialContext = NewSafeDialContext(&net.Dialer{
				Timeout:   15 * time.Second,
				KeepAlive: 30 * time.Second,
			})

			// Shallow copy the client
			safeClientCopy := *client
			safeClientCopy.Transport = safeTransport
			safeClient = &safeClientCopy
		}
	} else if client.Transport == nil {
		safeTransport := http.DefaultTransport.(*http.Transport).Clone()
		safeTransport.DialContext = NewSafeDialContext(&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		})
		safeClientCopy := *client
		safeClientCopy.Transport = safeTransport
		safeClient = &safeClientCopy
	}

	var resp *http.Response
	var err error

	for i := range maxRetries {
		if i > 0 && resp != nil {
			resp.Body.Close()
		}

		resp, err = safeClient.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				break
			}
			if !shouldRetry(resp.StatusCode) {
				break
			}
		}

		if i < maxRetries-1 {
			if err = sleepWithCtx(req.Context(), retryDelayUnit*time.Duration(i+1)); err != nil {
				if resp != nil {
					resp.Body.Close()
				}
				return nil, fmt.Errorf("failed to sleep: %w", err)
			}
		}
	}
	return resp, err
}

func sleepWithCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

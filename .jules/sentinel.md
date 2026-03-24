## 2025-02-28 - [Medium] Fix Missing HTTP Server Timeouts
**Vulnerability:** Go's standard `http.ListenAndServe` and unconfigured `http.Server` instances lack default timeouts for reading headers, reading bodies, and writing responses.
**Learning:** These default settings leave the application vulnerable to resource exhaustion and Denial of Service (DoS) attacks, such as Slowloris, because malicious clients can slowly send data and tie up server connections indefinitely.
**Prevention:** Always instantiate `http.Server` explicitly and set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and (optionally) `IdleTimeout` to reasonable values based on the expected request sizes and latencies.

## 2025-02-28 - [MEDIUM] Fix missing HTTP client timeouts in OAuth
**Vulnerability:** Go's standard `http.Post`, `http.PostForm`, `http.Get`, and the default `http.Client` do not have timeouts configured by default.
**Learning:** These defaults can leave the application vulnerable to resource exhaustion or indefinite hangs if the external service (like an OAuth provider) is slow, unresponsive, or experiencing an outage.
**Prevention:** Always instantiate `http.Client` explicitly with a sensible `Timeout` (e.g., `Timeout: 15 * time.Second`) before making outbound HTTP requests, instead of using the default package-level convenience functions.

## 2025-03-19 - [HIGH] Fix Timing Attack Vulnerability in Pico WebSocket Auth
**Vulnerability:** The Pico WebSocket authentication handler was using standard string equality (`==`) to compare the provided bearer token against the configured secret token.
**Learning:** String equality operators in Go return early as soon as a character mismatch is found. This allows an attacker to measure the time it takes for the server to reject the connection and iteratively guess the token character by character (a timing attack).
**Prevention:** Always use `subtle.ConstantTimeCompare` from the `crypto/subtle` package when comparing secrets, tokens, passwords, or cryptographic signatures to ensure the comparison time depends only on the length of the secret, not the contents.

## 2025-03-24 - [HIGH] Fix Length Timing Leak in subtle.ConstantTimeCompare Usage
**Vulnerability:** The `subtle.ConstantTimeCompare` function returns immediately if the lengths of the two byte slices differ, which can leak the exact length of the expected secret (token).
**Learning:** Even when using `subtle.ConstantTimeCompare` to avoid character-by-character timing leaks, length leaks can occur because the time taken varies if the attacker guesses an incorrect length compared to the correct length.
**Prevention:** Always wrap `subtle.ConstantTimeCompare` with a helper that handles length differences by executing a dummy comparison (`subtle.ConstantTimeCompare(actual, actual)`) to equalize the timing cost relative to the actual secret's length, before returning `false`.

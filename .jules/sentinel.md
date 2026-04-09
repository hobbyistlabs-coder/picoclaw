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

## 2026-04-09 - [Atomic Writes and Permissions for Session Storage]
**Vulnerability:** Google Messages session files (`session.json`) were being written using `os.WriteFile` with a permissive parent directory permission of `0755` (`0o755`).
**Learning:** Overly permissive directory permissions can allow other users to traverse and potentially list/read sensitive session tokens. Furthermore, using `os.WriteFile` allows for non-atomic writes which can leave session files in a corrupted, empty state on crashes, and does not restrict file creation permissions if an attacker pre-creates the file or places a symlink.
**Prevention:** Use `os.MkdirAll(dir, 0o700)` for sensitive local storage directories to restrict access to the owning user. Always use `fileutil.WriteFileAtomic` with restrictive file permissions (e.g. `0o600`) to guarantee writes are complete and strictly respect intended file attributes against race conditions.

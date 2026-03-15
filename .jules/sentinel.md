## 2025-02-28 - [Medium] Fix Missing HTTP Server Timeouts
**Vulnerability:** Go's standard `http.ListenAndServe` and unconfigured `http.Server` instances lack default timeouts for reading headers, reading bodies, and writing responses.
**Learning:** These default settings leave the application vulnerable to resource exhaustion and Denial of Service (DoS) attacks, such as Slowloris, because malicious clients can slowly send data and tie up server connections indefinitely.
**Prevention:** Always instantiate `http.Server` explicitly and set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and (optionally) `IdleTimeout` to reasonable values based on the expected request sizes and latencies.

## 2025-02-28 - [MEDIUM] Fix missing HTTP client timeouts in OAuth
**Vulnerability:** Go's standard `http.Post`, `http.PostForm`, `http.Get`, and the default `http.Client` do not have timeouts configured by default.
**Learning:** These defaults can leave the application vulnerable to resource exhaustion or indefinite hangs if the external service (like an OAuth provider) is slow, unresponsive, or experiencing an outage.
**Prevention:** Always instantiate `http.Client` explicitly with a sensible `Timeout` (e.g., `Timeout: 15 * time.Second`) before making outbound HTTP requests, instead of using the default package-level convenience functions.

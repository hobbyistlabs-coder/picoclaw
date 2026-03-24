## 2024-05-24 - Single Pass String Iteration for Token Estimation
**Learning:** `utf8.RuneCountInString(msg)` performs a full iteration over the string. If the string needs to be iterated over again (e.g., `for _, r := range msg`) to check specific rune properties, this results in two full passes over the string.
**Action:** Always count the total number of runes manually within the existing `for range` loop when a subsequent full iteration of the string is already required. This essentially halves the execution time.
## 2024-05-25 - Efficient String Building in Loops
**Learning:** In Go, string concatenation (`+=`) in a loop leads to $O(N^2)$ complexity due to immutability. Using `strings.Builder` provides $O(N)$ efficiency. Additionally, `fmt.Fprintf` has overhead due to format string parsing; direct `sb.WriteString` calls are significantly faster.
**Action:** Use `strings.Builder` for building strings in loops and prefer direct `WriteString` calls over `fmt.Fprintf` for maximum performance in hot paths.

## 2024-05-26 - Avoid Unnecessary `strings.ToLower`
**Learning:** Calling `strings.ToLower` on the entire message content allocates a new string and iterates over all runes. This causes measurable GC pressure and latency on hot paths like feature extraction during routing.
**Action:** Use fast paths to bypass `strings.ToLower`. For instance, check if a requisite character (like a dot `.`) exists, or check common casings directly (`DATA:IMAGE` vs `data:image`) before falling back to full case-normalization.
## 2025-03-20 - String Operations Fast Paths and Avoiding Double Searches
**Learning:** When trying to optimize `strings.ToLower`, ensure you don't introduce regressions with byte-to-rune casting on UTF-8 strings. Also, `strings.Contains(s, sub)` literally calls `strings.Index(s, sub)` under the hood. Using `strings.Contains` followed immediately by `strings.Index` to extract the position is an anti-pattern that searches the string twice, undermining the intended performance optimization.
**Action:** Always prefer a single `strings.Index` call over `Contains`+`Index`. Stick to one single optimization per PR to reduce risk and review burden.

## 2025-03-24 - Avoid strings.ToLower on large byte slices
**Learning:** Calling `strings.ToLower(string(body))` on large HTTP response bodies (up to 10MB) just to check the first few characters is extremely inefficient. It involves allocating a large string from the byte slice, and then allocating another large string for the lowercase conversion. This is an O(N) operation in terms of both time and memory.
**Action:** When performing prefix checks or substring searches on byte slices, avoid converting the entire slice to a string. Instead, use `bytes.HasPrefix` for exact matches, or bounded slice checks with `bytes.EqualFold` for case-insensitive matches (e.g., `bytes.EqualFold(body[:5], []byte("<html"))`).

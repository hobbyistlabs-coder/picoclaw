1. **Read PR Comments**: Use `read_pr_comments` to retrieve the comments on the pull request.
2. **Analyze Feedback**: Analyze the comments and determine what changes are required.
3. **Implement Changes**: Modify the code or configuration to address the feedback.
4. **Verify Fixes**: Run `go build ./...`, `go test ./...`, and `go generate ./...` if applicable.
5. **Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.**
6. **Submit**: Use `reply_to_pr_comments` to respond to the feedback and `submit` tool to update the pull request.

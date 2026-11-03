# AI usage

This project was developed with GitHub Copilot in VS Code.

- Tools used: workspace inspection, file reads, repository search, patch-based editing, PowerShell command execution, Go tests, race detector, vet, build, and live HTTP checks.
- AI helped understand the supplied assignment, complete the root Go service, restore the starter skeleton under `starter/`, diagnose the duplicate root declarations, fix the supplied debugging program, write tests, and draft documentation.
- An in-memory store was chosen over SQLite. This was changed from the initially considered SQLite approach to keep the submission dependency-free and aligned with the assignment's allowed choices; the durability and scaling trade-off is documented in `NOTES.md`.
- AI helped explain why global event-ID deduplication, campaign-scoped unique opens, UTC bucketing, and synchronized worker updates are separate correctness concerns.
- AI-generated code required debugging: the first batch-atomicity edit left an unused variable and caused compilation errors; the issue was corrected and verified with `go test`, `go test -race`, `go vet`, and `go build`.
- The original supplied files are now present and were used for the debugging fixes. No claim is made about work outside the repository.

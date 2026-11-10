# Initial sprint interface review

**English** · [Español](es/interface-check.md) · [Italiano](it/interface-check.md)

The initial review covered the shared Bubble Tea v2/Bubbles v2 TUI: monospaced text, separators and written status, without CSS or custom animation. It is not a complete visual/accessibility review of Windows/macOS/Linux terminals.

| Category | Evidence | Result |
| --- | --- | --- |
| Typography | `ansi.Wrap/Truncate` and terminal width | Wrapping implemented; manual visual review pending |
| Surfaces | Text-separated header, tabs, history and input | No extra visual dependencies |
| Animation | Only the standard cursor | No custom 10%-speed playback review applies |
| Icons | Written connection/ACK state and bracketed selection | Status does not depend only on color/icons |
| Performance | 500 displayed messages and 100 per page | Bounded history memory; load profiling pending |

| Severity | Location | Before | After | Reason |
| --- | --- | --- | --- | --- |
| Medium | `pkg/tui/model.go`, send/reload | Reload could discard a pending ID | Reload waits for confirmation/cancellation | Prevent ambiguous retries |
| Medium | `pkg/tui/model.go`, input | Out-of-limit input lacked explanation | Byte/control-character limit shown | Make the next action visible |

Color effects for connection, frequent message animation and a new component style system were rejected as unnecessary for the first trial; text works across terminals and existing Bubbles components cover the flow.

A real SSH PTY rendered the prototype header and persisted messages; transport tests check control sequences. Screen readers, zoom, 12-column terminals and visual behavior across all native terminals remain unreviewed. These are documented alpha limitations. The later native E2EE trial is recorded in [verification](verification.md).

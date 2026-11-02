# Experimental alpha verification

**English** · [Español](es/verification.md) · [Italiano](it/verification.md)

Reproducible evidence for local/LAN code, not an independent audit or a physical-machine network trial. VPS/domain/provider work belongs to future hosting. Per-commit results are in [GitHub Actions](https://github.com/karatekah/karatty/actions/workflows/ci.yml).

| Area | Automated evidence |
| --- | --- |
| Email-free signup/recovery | `TestAnonymousSignupAndOfflineRecovery`: possession proof, single-use challenge, wrong code, revocation and code rotation |
| Consent/key access | `TestE2EEKeyDistributionRequiresAcceptedDirect`: no identities/one-time claims before acceptance; invalid signatures rejected; consumed keys never reactivate |
| E2EE messages | `TestE2EEClientExchangeAndLocalPassphrase`: matching self/peer fingerprints, send/reply, opaque PostgreSQL body, 4096 bytes, restart, altered ACK, tampered envelope and stable retry |
| Device recovery | Same integration replaces SSH key/Olm identity, requires re-verification, migrates readable archive, retains uncertain sends and checks old-state deletion |
| Closure/deletion | Closed-chat local history remains readable; new sends fail; account files are removed. Server suites cover authorization/deletion |
| Bounded forward secrecy | Rust advances ratchets and shows current account/session copies cannot decrypt a consumed message; replay/tampering fail; out-of-order messages work. Unlocked archives/old snapshots are outside this guarantee |
| Local archive | `pkg/localstate`: wrong phrase, tampering, process lock, replacement and 16 MiB reopen. `pkg/secureclient`: 720-hour cleanup and history ordering when ACK precedes backlog |
| Transport bounds | `TestMaximumEnvelopePageOverHTTPAndWebSocket`: 100 maximum-size, double-escaped envelopes without truncation |
| TUI | `pkg/tui`: complete fingerprints, external-comparison instruction and saved-send retry with the same ID |
| Native client | `TestE2EEClientNativeTUI`: two real PTY executables unlock archives, share/verify fingerprints and deliver a message stored as an opaque envelope |
| Local/LAN TLS | `scripts/TestDevelopmentCertificateLAN`: trust chain, localhost, IPv4/IPv6 loopback/private-IP SAN; rejects absent IP, overwrite, invalid/unspecified/multicast addresses |
| Distribution | CI builds/extracts/runs version and Go–Rust bridge on Windows/Linux/macOS; checks hashes, licenses and source commit |
| Service | Real PostgreSQL, TLS/WSS, executable client and legacy gateway; race detector, `go vet`, `govulncheck`, Cargo.lock/RustSec audit, Docker, shell syntax and systemd verification |

## Windows trial, November 1, 2026

`TestE2EEClientExchangeAndLocalPassphrase` passed using native PostgreSQL 17.11 and `karatty-crypto.exe` from candidate `d39b5da`. It exercises two clients over loopback TLS, including exchange, history, recovery and retries. That integration test itself is not a manual two-TUI Windows session. The automated native-PTY test above runs on Linux in CI.

`go vet ./...`, `go test -count=1 ./...` with database/bridge configured, `go build ./cmd/...`, certificate tests and relative Markdown-link checks also passed. Compilation alone does not prove antivirus will permit execution.

Initially, Defender blocked `.local/lan-alpha/server.exe` when attempted on the private IP, classifying it as `Behavior:Win32/DefenseEvasion.A!ml`. No protection was disabled or exclusion added. No false positive was established, so that IP-listener trial did not pass.

### Native trial after restoration

The user later restored `.local/defender-review/server.exe` from quarantine. The `4c0356f` build has SHA-256 `9a81de63cc5149d551bcc4abc3cf0b6a8b83b8658a58d6fc819bb8fe7db830f1`. Its `--help` execution and native startup at `https://127.0.0.1:28443` then passed in E2EE mode without SSH gateway.

The Windows `v0.1.0-alpha.1` client package from that commit was extracted. Two real TUI processes ran through ConPTY on the same machine: distinct-key registration/local passphrases, direct request/acceptance, sharing self fingerprints and verifying both contacts. Alice sent and Bob replied; both messages appeared in the other client and received ACK sequences 1/2. Closing/reopening both TUIs recovered both messages and synchronized through sequence 2.

PostgreSQL held two E2EE-shaped envelopes without either readable test message. The texts were also absent from operational server logs. Clients and the trial server were stopped afterward.

Defender real-time/behavior protection stayed active with signatures `1.463.83.0`; no new detection for that executable was observed during the trial. This validates that restored build on that device, not a false-positive determination or compatibility with other Windows installations. No samples were sent to Microsoft.

Only one machine was available. **Two physical LAN hosts remain untested.** [Setup instructions](quickstart.md) and certificate tests do not prove cross-machine reachability.

## Evidence limits

- No external Karatty audit. A vodozemac audit covers that library/audited version, not automatically this integration or later versions.
- Persistence-before-send/display, locking and replacement are tested; physical power loss across all three filesystems is not.
- Sequence, time, account name and service availability are server-controlled metadata. The encrypted payload binds conversation/sender/client ID; a server can still hide messages or deny service.
- Archive plaintext is bounded to 16 MiB. On exhaustion, saving fails before transmitting an unsaved advanced ratchet. Cleanup needs a running client and cannot erase offline devices/other snapshots.
- Native-TUI screen-reader/accessibility and all-platform visual reviews remain incomplete; CI validates behavior/binaries, not that experience.
- Real hosted-service installation, provider backup expiry/restore, domain/certificate checks and operator policies/contact remain future work. Record them under [release](release.md) before announcing hosting.

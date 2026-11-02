# Publishing Karatty v0.1.0-alpha.1

**English** · [Español](es/release.md) · [Italiano](it/release.md)

The [first alpha](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1) distributes source, documentation and client packages for local use with your own server. It is experimental. LAN configuration is documented, but a two-physical-host conversation is unverified. No project-hosted Internet service is available.

## Scope and evidence

| Component | Status |
| --- | --- |
| E2EE client, recovery, retries, retention | Implemented; [evidence and limits](verification.md) |
| Windows trial | Native TLS server/two TUIs: signup, acceptance, fingerprints, messages and reopened history |
| Defender | Trial build required quarantine restoration. Protection stayed active; false positive and other-installation behavior remain undetermined |
| Packages | Windows x64, Linux x64, macOS Apple Silicon; client/bridge, licenses, locks, BUILD_INFO/hashes. Build server separately |
| Two physical LAN hosts | Pending; configuration is experimental |
| Domain/VPS/providers | Future work, not required for local alpha |
| Future hosted-service privacy/terms | Inactive drafts; administrators disclose their own instance/copies |
| Security reports | Private GitHub channel under [SECURITY.md](../SECURITY.md) |

Own code is Apache-2.0; dependencies retain their licenses. Downloads are unsigned/not notarized. Karatty has no independent audit.

## Publication checks

1. Record CI for the tagged commit: PostgreSQL, E2EE, native TUI, Go/Rust, races, package extraction and dependency advisories. The Windows manual evidence identifies its exact build; it does not automatically cover another binary/device.
2. Review full history/current files for personal information, exposed credentials, database exports and private keys. Review retained Actions logs before they become public. Keep `.local/`, `.env`, `bin/`, `dist/` ignored.
3. Review README, installation, architecture, security, contribution/changelog pages. Keep Defender/LAN limitations visible and future hosted-policy drafts explicitly inactive.
4. Merge preparation into `main`, preserve conventional commits and verify CI on the final commit. Download its packages; verify SHA-256, licenses and `BUILD_INFO` source commit.
5. Tag `v0.1.0-alpha.1` at that commit, attach three packages and a combined `SHA256SUMS`, and write evidence/limitations in the notes.
6. Make the repository public, enable private vulnerability reports, publish as a **prerelease**, and check repository, README, tag, Release/downloads without authentication.

LAN validation requires a two-machine trial before claiming that support verified. Domain/VPS/mailboxes/provider tests belong to [future hosting](operations.md).

## Build client packages

On the target platform with PowerShell, Go 1.27.3 and Rust 1.99.1:

```powershell
./scripts/build-release.ps1 -Version v0.1.0-alpha.1
```

The script builds native `karatty` without CGo and Rust `karatty-crypto`. It creates a directory/archive in `dist/VERSION/`, binary/archive hashes, own/third-party licenses, locks and source-commit `BUILD_INFO`. Existing output directories cannot be overwritten. Keys, private configuration and databases are excluded. Build the server separately under [quickstart](quickstart.md).

CI builds/extracts/executes each bridge on Windows/Linux/macOS. Actions packages are Release inputs, not a substitute for publication/access checks. [Installation](install.md) covers verification, paired executables, identity and chat.

## Documentation after publication

English is now primary with complete [Spanish](es/release.md)/[Italian](it/release.md) counterparts; see [translation maintenance](translations.md). The repository is now `karatekah/karatty`. Published alpha.1 binaries, checksums, source commit `5e4adde` and tag remain immutable. Its Go module used the former name; current development uses the new one. [Version notes](releases/v0.1.0-alpha.1.md) describe the published artifacts.

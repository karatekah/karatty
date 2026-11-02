# Install and use Karatty alpha

**English** · [Español](es/install.md) · [Italiano](it/install.md)

This alpha targets local use with a server you administer. No project-hosted Internet service is available; domain/VPS are future work. Download [v0.1.0-alpha.1](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1). Use the [local/LAN guide](quickstart.md) to start a server and try two accounts. Two-physical-machine LAN testing remains pending.

These examples use a same-machine server. For LAN, replace the URL in **every** command with the private server address and `CA_PATH` with the `ca.crt` received/verified with its administrator. Keep the same URL to retain local identity. Client machines do not need PostgreSQL.

## Download and verify

Client packages contain both executables; build the server from source using the local guide. Other architectures have no prebuilt package in this release.

| Platform | File |
| --- | --- |
| Windows x64 | `karatty-v0.1.0-alpha.1-windows-amd64.zip` |
| Linux x64 | `karatty-v0.1.0-alpha.1-linux-amd64.tar.gz` |
| macOS Apple Silicon | `karatty-v0.1.0-alpha.1-darwin-arm64.tar.gz` |

Also download the same release's `SHA256SUMS` and compare before extraction:

```powershell
# Windows: compare the result with the downloaded archive line.
Get-FileHash .\karatty-v0.1.0-alpha.1-windows-amd64.zip -Algorithm SHA256
```

```sh
# Linux
sha256sum -c SHA256SUMS --ignore-missing
# macOS: compare the result with SHA256SUMS.
shasum -a 256 karatty-v0.1.0-alpha.1-darwin-arm64.tar.gz
```

Extract into your own directory and keep `karatty`/`karatty-crypto` together (`.exe` on Windows). Optionally add that directory to PATH. Licenses, dependency locks and source-commit `BUILD_INFO` are included. Alpha packages are unsigned/not notarized. Hashes check integrity against the published manifest; they do not independently authenticate the author.

Windows Defender blocked one server build during testing. Its owner restored it and local tests then passed with protection active; the classification has not been confirmed wrong. If antivirus blocks a file, retain the alert and check provenance rather than disabling protection. [Verification](verification.md) identifies the tested build and limits.

## Create an identity

Use a dedicated Ed25519 SSH key. OpenSSH asks for its protecting passphrase; choose a separate passphrase for the local message archive.

```sh
ssh-keygen -t ed25519 -f YOUR_KEY_PATH
karatty register --server https://127.0.0.1:8443 --ca-cert CA_PATH --key YOUR_KEY_PATH --username YOUR_ALIAS --public
```

In PowerShell, use `./karatty.exe` from the extracted folder. Usernames allow ASCII letters, digits, hyphens/underscores, up to 64 characters. Registration asks for no email or real name. Public keys and account/conversation metadata are stored; a pseudonym is not absolute anonymity.

Store the recovery code shown once **outside this device**. Then choose/confirm a local passphrase of at least 12 characters. It never leaves the device, and the recovery code cannot replace it. Runtime prompts are currently Spanish.

## Open a direct chat

```sh
karatty chat --server https://127.0.0.1:8443 --ca-cert CA_PATH --key YOUR_KEY_PATH
```

Enter the SSH-key passphrase if protected, then the local archive passphrase. In the TUI:

1. Alice requests `/direct bob`.
2. Bob checks `/requests` and accepts with `/direct alice`.
3. Each runs `/myfingerprint` and shares that own fingerprint over another trusted channel.
4. Alice checks `/fingerprint bob`, compares Bob's independently supplied value and uses `/verify bob BOB_FINGERPRINT`. Bob does the equivalent for Alice.
5. Type a message and press Enter. Tab switches chat; PgUp/PgDn scroll; Ctrl+L reloads; Ctrl+C exits.

Independent comparison detects first-contact server identity substitution. Copying a fingerprint solely through the same conversation is not an independent check. A changed identity stops sending.

ACK means the server persisted the encrypted envelope, not that the recipient read it. Ctrl+R retries an unconfirmed send with the same ID/envelope. After restarting, `/pending` lists saved sends for the direct chat and `/retry ID` retries one. CLI equivalents are `pending` and `retry-pending --conversation CHAT_ID --id SEND_ID`. Esc returns to editing while preserving the saved send for review. Discarding an uncertain send cannot retract a copy already accepted by the server. `/decline USER`, `/block USER` and `/unblock USER` manage requests/blocking.

## History, recovery and deletion

Messages are retained for 720 hours (30 days). The server hides expired messages and purges at startup/hourly. The client cleans its archive on open/save and filters expired reads; offline devices cannot run cleanup. The local guide sets up no automatic backups. Administrators must disclose/verify their retention. The future hosted-service requirement is at most seven additional backup days, not an active service promise. Forward secrecy does not protect an unlocked copy or compromised device.

If your SSH key is lost, generate another and run:

```sh
karatty recover --server https://127.0.0.1:8443 --ca-cert CA_PATH --key NEW_KEY --username YOUR_ALIAS
```

Enter the saved code and retain its replacement. Open the client with the new key. If the old archive remains locally, close other clients and use the same local passphrase to migrate readable history. Old state is removed only after the new state is saved. Earlier sends without ACK are listed by `karatty uncertain` and do not resend automatically. Recovering the account cannot restore a lost archive or passphrase.

Contacts must compare the new fingerprint through another channel and run `/reverify USER FINGERPRINT`. Pending sends require explicit `/discardpending USER` before accepting the new identity; check whether they arrived before repeating them. Previously undecrypted messages may be unavailable.

To delete the account:

```sh
karatty delete-account --server https://127.0.0.1:8443 --ca-cert CA_PATH --key YOUR_KEY_PATH --confirm-delete
```

This removes the account, keys, sessions and authored messages from active storage and that account's local files under this device's configured state root. Close other clients so cleanup can acquire its lock. Other people's copies/devices cannot be remotely erased; backups follow their own expiry. If server deletion succeeds but local cleanup fails, the CLI reports the path requiring review.

## Update or uninstall

Close Karatty, verify/extract the new package and replace **both** executables together. Keep SSH keys and state. The alpha supports one device per account.

Default state lives under the user's configuration directory: `%AppData%\karatty` (Windows), `$XDG_CONFIG_HOME/karatty` or `~/.config/karatty` (Linux), `~/Library/Application Support/karatty` (macOS). Override with `--state-dir ABSOLUTE_PATH`. Removing executables does not delete a server account; run `delete-account` first if that is intended.

See [architecture](architecture.md), [E2EE guarantees/limits](adr/0003-e2ee-1to1.md) and [contributing](../CONTRIBUTING.md).

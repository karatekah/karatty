# Karatty on your machine or local network

**English** · [Español](es/quickstart.md) · [Italiano](it/quickstart.md)

The alpha uses a server on your machine/LAN, PostgreSQL and local TUI clients. No domain, VPS, email or Internet exposure is required. Both participants must reach the same server. Each account uses one device.

Use `--public-signup`: pseudonymous email-free registration, 1:1 chats and E2EE envelopes. The flag does not require publishing an Internet service. Do not use `--legacy`, invitations, channels or the SSH gateway for this alpha. The TUI currently displays Spanish prompts; the documentation language switch does not translate runtime messages.

## 1. Build or install

Building requires Go 1.27.3, Rust 1.99.1 and OpenSSH. Only the server needs PostgreSQL 17.11, native or Docker. Clients may use both executables from a [verified package](install.md) without installing Go, Rust or PostgreSQL. Client packages exclude the server.

Clone with `git clone https://github.com/karatekah/karatty.git` and enter `cd karatty`. Alternatively download tagged source from the Release. The current module is `github.com/karatekah/karatty`; historical alpha.1 source retains its old module name but builds locally from its own consistent source.

PowerShell from the repository root:

```powershell
New-Item -ItemType Directory -Force bin,.local | Out-Null
go build -trimpath -o bin/server.exe ./cmd/server
go build -trimpath -o bin/karatty.exe ./cmd/termlink
cargo build --locked --release --manifest-path crypto/Cargo.toml
Copy-Item crypto/target/release/karatty-crypto.exe bin/karatty-crypto.exe
```

Linux/macOS:

```sh
mkdir -p bin .local
chmod 700 .local
go build -trimpath -o bin/server ./cmd/server
go build -trimpath -o bin/karatty ./cmd/termlink
cargo build --locked --release --manifest-path crypto/Cargo.toml
cp crypto/target/release/karatty-crypto bin/karatty-crypto
```

The remaining examples use PowerShell. On Linux/macOS use `export VARIABLE=value` and executables without `.exe`; flags/TUI commands are identical.

## 2. Start PostgreSQL

With native PostgreSQL, create a dedicated database and set `DATABASE_URL`. Never test against production. The server applies migrations at startup:

```powershell
$env:DATABASE_URL = 'postgres://USER:URL_ENCODED_PASSWORD@127.0.0.1:5432/karatty?sslmode=disable'
```

URL-encode reserved password characters. Here `sslmode=disable` applies only to a database on the same machine. Keep credentials out of Git and public captures.

Alternatively, development Compose creates the `termlink` database and exposes PostgreSQL only on loopback:

```powershell
$env:TERMLINK_DB_PASSWORD = 'termlink-local-development-only'
docker compose -f deployments/compose.yml up -d db
$env:DATABASE_URL = "postgres://termlink:$env:TERMLINK_DB_PASSWORD@127.0.0.1:5432/termlink?sslmode=disable"
```

The password is a local-development example. Do not start Compose's `app` profile: it retains the invitation/SSH prototype for development tests.

## 3. Choose local or LAN with TLS

**This machine only:**

```powershell
go run ./scripts/devcert.go
./bin/server.exe --addr 127.0.0.1:8443 --tls-cert .local/tls/server.crt --tls-key .local/tls/server.key --public-signup
```

Both clients use `https://127.0.0.1:8443` and `.local/tls/ca.crt`. Keep the server terminal open.

**LAN:** find the server's private IPv4 with `ipconfig` (Windows) or `ip addr` (Linux). Replace `192.168.1.50` in both commands with that address:

```powershell
go run ./scripts/devcert.go --out .local/tls-lan --ip 192.168.1.50
./bin/server.exe --addr 192.168.1.50:8443 --tls-cert .local/tls-lan/server.crt --tls-key .local/tls-lan/server.key --public-signup --public-network
```

`--public-network` permits non-loopback listening, including LAN, and requires TLS/E2EE without `--dev`. Both clients use `https://192.168.1.50:8443`, including a client running on the server machine. Choose a stable URL before registering: encrypted state is bound to it; changing it can create a new local identity. A DHCP reservation helps keep the address.

Give participants **only `ca.crt`** through a trusted channel and compare its SHA-256 outside Karatty (`Get-FileHash PATH/ca.crt -Algorithm SHA256` or `sha256sum PATH/ca.crt`). `--ca-cert` trusts it only for this client, without global installation. `server.key` stays on the server; each participant keeps a distinct SSH key/state. Never share state directories between machines.

The server firewall must allow TCP 8443 from clients in the private subnet. Limit any rule to that subnet/private profile; keep the firewall active. Guest networks may isolate devices. Do not expose PostgreSQL or forward router ports. Another Windows machine can check reachability with `Test-NetConnection 192.168.1.50 -Port 8443`.

Development certificates expire after 30 days; the CA private key is never saved. Existing certificates cannot be overwritten by the generator. To renew or change IP, use a new `--out`, restart with new files and redistribute/recheck the CA. Never disable TLS validation.

## 4. Register two participants

In another terminal, Alice generates a dedicated key and registers:

```powershell
ssh-keygen -t ed25519 -f .local/alice
./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice --username alice --public
./bin/karatty.exe chat --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice
```

Bob uses another terminal/key:

```powershell
ssh-keygen -t ed25519 -f .local/bob
./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/bob --username bob --public
./bin/karatty.exe chat --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/bob
```

For LAN, replace `--server` and `--ca-cert` in **every** command with the LAN URL and received CA path. Each person generates their key locally rather than copying a contact's key. Create the key's directory first if needed. With an extracted package, run `./karatty.exe` beside `karatty-crypto.exe`.

Store the one-time-displayed recovery code outside the device. Choose/confirm a local passphrase independent of the SSH-key passphrase. Unlock the archive whenever opening chat. Separate keys isolate Alice/Bob state even on one machine. Optional `--state-dir ABSOLUTE_PATH` must be repeated consistently on each invocation.

## 5. Accept, verify and chat

Inside both TUIs:

```text
Alice: /direct bob
Bob:   /requests
Bob:   /direct alice
Both: Ctrl+L to reload; Tab to select the direct chat.
Both: /myfingerprint
       Share and compare fingerprints through another trusted channel.
Alice: /fingerprint bob
Alice: /verify bob BOB_FINGERPRINT
Bob:   /fingerprint alice
Bob:   /verify alice ALICE_FINGERPRINT
Alice: Hello Bob
Bob:   Hello Alice
```

Check readable delivery in the other client and ACK at the sender. Close both with Ctrl+C, reopen with the same keys, URL and passphrases, and check history. ACK means persistence, not reading. `/pending` and `/retry ID` inspect/retry unconfirmed sends with the same envelope.

To validate LAN, repeat on **two physical machines**. Two processes on one host connected to its private IP do not prove remote firewall/Wi-Fi reachability. Record operating systems, terminals and results without private keys, codes or conversations. See [existing evidence](verification.md) and [recovery/deletion](install.md).

## Stop and retain data

Ctrl+C stops clients/server. If using Docker, `docker compose -f deployments/compose.yml stop db` stops PostgreSQL. Keep the database, keys and encrypted state to reuse accounts. Do not use `down -v` if retaining the database.

E2EE mode hides messages after 720 hours and purges at startup/hourly. Deduplication also lasts 30 days. Local cleanup runs only while the client is used; offline devices and external copies cannot be remotely erased. This guide sets up no automatic backups. Administrators must disclose/verify copy retention; [backup tools](operations.md) require separate configuration.

## Check development

```powershell
$env:TERMLINK_TEST_DATABASE_URL = $env:DATABASE_URL
$env:KARATTY_CRYPTO_BIN = (Resolve-Path bin/karatty-crypto.exe).Path
./scripts/check.ps1
```

Integration tests create/delete random schemas; the role needs permission to create them. Missing variables skip the corresponding tests. On Linux/macOS export the same variables and follow [contribution checks](../CONTRIBUTING.md). Internet hosting is covered separately in the [future deployment plan](deploy.md).

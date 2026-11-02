# Future Internet hosting

**English** · [Español](es/deploy.md) · [Italiano](it/deploy.md)

This procedure is future work. The [local/LAN alpha](quickstart.md) needs none of these purchases. No VPS deployment or project-hosted service exists. Hosting requires the selected `karatty.dev` domain, administrative VPS access, a private bucket and public operator contact. Do not announce hosted registration before [operational checks](operations.md).

## Planned configuration

A 1-vCPU/2-GiB Ubuntu 24.04 LTS instance runs PostgreSQL 17 and Go on one machine. Ubuntu 24.04 has [standard maintenance through May 2029](https://ubuntu.com/about/release-cycle). Clients connect directly by TLS to `https://chat.karatty.dev`, without a proxy hiding client IPs from rate limiting. Do not install the messaging SSH gateway. Administrative SSH is restricted to operator keys.

Create an unproxied DNS A record for `chat.karatty.dev`. Publish AAAA only after testing IPv6 listening/firewall. Allow TCP 443 for API and 80 for Certbot HTTP challenges; restrict TCP 22 for administration. PostgreSQL must use loopback/SCRAM without a public port.

## Database and service account

Install PostgreSQL 17/tools from the [official Ubuntu repository](https://www.postgresql.org/download/linux/ubuntu/). Check `postgres --version`/`pg_dump --version` and use matching major versions for backup/restore. Create a non-superuser `karatty` role without role-creation privileges and its owned `karatty` database. Set its password through `\password karatty` in psql rather than shell history. It needs schema/migration creation permissions inside its own database.

```sh
sudo useradd --system --home /var/lib/karatty --shell /usr/sbin/nologin karatty
sudo install -d -o root -g karatty -m 0750 /etc/karatty /etc/karatty/tls
sudo install -d -o root -g root -m 0755 /opt/karatty
sudoedit /etc/karatty/server.env
sudo chmod 600 /etc/karatty/server.env
```

The environment file contains one variable; URL-encode reserved password characters:

```dotenv
DATABASE_URL=postgres://karatty:URL_ENCODED_PASSWORD@127.0.0.1:5432/karatty?sslmode=disable
```

`sslmode=disable` is only for the database on that same VPS, not a remote database. Keep environment files/private certificates out of Git.

## Binary and TLS

From a CI-verified commit with Go 1.27.3, build `go build -trimpath -o bin/karatty-server ./cmd/server`. Install it and the supplied unit/hook:

```sh
sudo install -m 0755 bin/karatty-server /opt/karatty/server
sudo install -m 0644 deployments/systemd/karatty.service /etc/systemd/system/karatty.service
sudo install -d -m 0755 /etc/letsencrypt/renewal-hooks/deploy
sudo install -m 0755 deployments/renew-certificate.sh /etc/letsencrypt/renewal-hooks/deploy/karatty
```

Install Certbot and issue a standalone certificate for `chat.karatty.dev` after checking DNS/port 80. The deploy hook copies cert/key with `root:karatty 0640` and acts only on this domain. If the first issuance did not run it, invoke with `RENEWED_DOMAINS=chat.karatty.dev` and `RENEWED_LINEAGE` set to Certbot's new directory. Test renewal before proceeding. Restart reloads the certificate in memory; clients reconnect from their sequence.

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now karatty.service
sudo systemctl status karatty.service
curl --fail https://chat.karatty.dev/healthz
```

The unit requires `--public-signup --public-network` plus TLS cert/key. It rejects channels, plaintext-shaped messages and a remote SSH TUI, and activates startup/hourly purging. Delivery metadata remain visible. Envelope structure cannot prove a modified client encrypted correctly; tests refer to the official client.

## Backup and restore

Install `age`, `rclone`, `pg_dump`. Create a dedicated private R2 bucket and bucket-scoped credential. Store `rclone.conf` at `/etc/karatty/rclone.conf`, `root:karatty 0640`. Configure `/etc/karatty/backup.env` (0600, read by systemd):

```dotenv
BACKUP_RECIPIENT=age1PUBLIC_RECIPIENT
BACKUP_REMOTE=r2:PRIVATE_BUCKET/db-backups
RCLONE_CONFIG=/etc/karatty/rclone.conf
```

Keep the private `age` identity outside the VPS. Install scripts/units:

```sh
sudo install -m 0755 scripts/backup.sh /opt/karatty/backup.sh
sudo install -m 0644 deployments/systemd/karatty-backup.service deployments/systemd/karatty-backup.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start karatty-backup.service
sudo systemctl enable --now karatty-backup.timer
sudo systemctl list-timers karatty-backup.timer
```

Apply three-day remote expiry and verify actual removal under [operations](operations.md). Use `scripts/restore-check.sh` on a machine holding the private identity, restoring only into an empty `termlink_restore_…` database. Check ciphertext, migrations, registration and authorization afterward. Never restore over the active database.

## Before announcing hosting

From another network, install each package and try two fresh accounts: offline recovery codes, acceptance, independent fingerprint comparison, bidirectional exchange, disconnect/restart, blocking, recovery and deletion. Confirm opaque API/database envelopes, hidden expired messages and physical row removal. Test backup/restore and remote expired-object deletion. Complete operator identity/contact/policies and record dated results under [release](release.md).

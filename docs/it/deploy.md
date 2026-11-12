# Futuro hosting Internet

[English](../deploy.md) · [Español](../es/deploy.md) · **Italiano**

Procedura per una fase futura. L'[alpha locale/LAN](quickstart.md) non richiede questi acquisti. Nessun VPS o servizio del progetto è operativo. Occorrono dominio `karatty.dev`, accesso amministrativo, bucket privato e contatto pubblico dell'operatore. Non annunciare registrazioni ospitate prima dei [controlli operativi](operations.md).

## Configurazione prevista

Istanza 1 vCPU/2 GiB Ubuntu 24.04 LTS, PostgreSQL 17 e Go sulla stessa macchina. Ubuntu 24.04 ha [manutenzione standard fino a maggio 2029](https://ubuntu.com/about/release-cycle). TLS diretto a `https://chat.karatty.dev`, senza proxy che nasconda IP ai limiti. Nessun gateway SSH di messaggistica; SSH amministrativo limitato alle chiavi dell'operatore.

DNS A senza proxy per `chat.karatty.dev`; pubblicare AAAA solo dopo prova IPv6/firewall. TCP 443 API, 80 sfide Certbot, 22 ristretto all'amministrazione. PostgreSQL su loopback con SCRAM, senza porta pubblica.

## Base e utente del servizio

Installa PostgreSQL 17/strumenti dal [repository ufficiale Ubuntu](https://www.postgresql.org/download/linux/ubuntu/). Verifica `postgres --version`/`pg_dump --version` e stessa versione principale per backup/ripristino. Crea ruolo `karatty` non superuser, senza creazione ruoli, e base propria. Imposta password con `\password karatty` in psql, non nella cronologia shell. Il ruolo deve creare schema/migrazioni nella propria base.

```sh
sudo useradd --system --home /var/lib/karatty --shell /usr/sbin/nologin karatty
sudo install -d -o root -g karatty -m 0750 /etc/karatty /etc/karatty/tls
sudo install -d -o root -g root -m 0755 /opt/karatty
sudoedit /etc/karatty/server.env
sudo chmod 600 /etc/karatty/server.env
```

File ambiente con una sola variabile e password codificata per URL:

```dotenv
DATABASE_URL=postgres://karatty:PASSWORD_CODIFICATA@127.0.0.1:5432/karatty?sslmode=disable
```

`sslmode=disable` vale solo per la base sullo stesso VPS, non per database remoti. Niente file ambiente o certificati privati in Git.

## Binario e TLS

Da commit con CI verde e Go 1.27.4: `go build -trimpath -o bin/karatty-server ./cmd/server`. Installa binario, unità e hook:

```sh
sudo install -m 0755 bin/karatty-server /opt/karatty/server
sudo install -m 0644 deployments/systemd/karatty.service /etc/systemd/system/karatty.service
sudo install -d -m 0755 /etc/letsencrypt/renewal-hooks/deploy
sudo install -m 0755 deployments/renew-certificate.sh /etc/letsencrypt/renewal-hooks/deploy/karatty
```

Installa Certbot e certificato standalone per `chat.karatty.dev`, dopo DNS/porta 80. L'hook copia certificato/chiave con `root:karatty 0640`, solo per quel dominio. Se non eseguito alla prima emissione, invocalo con `RENEWED_DOMAINS=chat.karatty.dev` e `RENEWED_LINEAGE` sulla nuova directory Certbot. Prova rinnovo prima di procedere. Il riavvio ricarica il certificato; i client riprendono dalla sequenza.

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now karatty.service
sudo systemctl status karatty.service
curl --fail https://chat.karatty.dev/healthz
```

L'unità richiede `--public-signup --public-network` e TLS. Rifiuta canali, corpi leggibili/TUI SSH remota e abilita pulizia all'avvio/ogni ora. Restano metadati di consegna. La struttura non dimostra che client modificati cifrino correttamente; i test riguardano il client ufficiale.

## Backup e ripristino

Installa `age`, `rclone`, `pg_dump`. Configura bucket R2 dedicato privato/credenziale limitata. Salva `rclone.conf` in `/etc/karatty/rclone.conf`, proprietario `root:karatty`, permessi 0640. Configura `/etc/karatty/backup.env` (0600, letto da systemd):

```dotenv
BACKUP_RECIPIENT=age1DESTINATARIO_PUBBLICO
BACKUP_REMOTE=r2:BUCKET_PRIVATO/db-backups
RCLONE_CONFIG=/etc/karatty/rclone.conf
```

Identità privata age fuori dal VPS. Installa script/unità:

```sh
sudo install -m 0755 scripts/backup.sh /opt/karatty/backup.sh
sudo install -m 0644 deployments/systemd/karatty-backup.service deployments/systemd/karatty-backup.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start karatty-backup.service
sudo systemctl enable --now karatty-backup.timer
sudo systemctl list-timers karatty-backup.timer
```

Imposta scadenza remota a tre giorni e verifica rimozione effettiva secondo [operazioni](operations.md). Da computer con identità privata, usa `scripts/restore-check.sh` solo su base vuota `termlink_restore_…`. Controlla cifrato, migrazioni, registrazione/autorizzazione. Non ripristinare sulla base attiva.

## Prima dell'annuncio

Da altra rete, installa ogni pacchetto e prova due account nuovi: codici offline, accettazione, impronte indipendenti, messaggi bidirezionali, disconnessione/riavvio, blocco, recupero, eliminazione. Verifica involucri opachi in API/base, scaduti nascosti e righe eliminate. Prova backup, ripristino e rimozione oggetti scaduti. Completa identità/contatto/politiche dell'operatore e registra data/risultati in [pubblicazione](release.md).

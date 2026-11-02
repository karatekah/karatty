# Karatty sul proprio computer o in LAN

[English](../quickstart.md) · [Español](../es/quickstart.md) · **Italiano**

L'alpha usa server locale/LAN, PostgreSQL e TUI locali. Non richiede dominio, VPS, email o esposizione Internet. Entrambi i partecipanti devono raggiungere lo stesso server. Un dispositivo per account.

Usa `--public-signup`: pseudonimi senza email, diretti 1:1 e involucri E2EE. Il nome del flag non obbliga a esporre il servizio su Internet. Non usare `--legacy`, inviti, canali o gateway SSH per l'alpha. I messaggi runtime della TUI sono ancora spagnoli; il selettore cambia solo la documentazione.

## 1. Compilare o installare

Per compilare servono Go 1.27.3, Rust 1.99.1 e OpenSSH. Solo il server necessita PostgreSQL 17.11, nativo o Docker. I client possono usare i due eseguibili di un [pacchetto verificato](install.md), senza Go/Rust/PostgreSQL. Il server non è incluso.

Clona con `git clone https://github.com/karatekah/karatty.git`, poi `cd karatty`, oppure scarica il sorgente della release. Il modulo attuale è `github.com/karatekah/karatty`; lo storico alpha.1 conserva il vecchio nome e si compila dal proprio sorgente coerente.

PowerShell dalla radice:

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

Gli esempi successivi usano PowerShell. Su Linux/macOS usa `export VARIABILE=valore` ed eseguibili senza `.exe`; flag e comandi TUI sono uguali.

## 2. Avviare PostgreSQL

Con installazione nativa, crea una base dedicata e imposta `DATABASE_URL`. Non usare produzione. Il server applica migrazioni all'avvio:

```powershell
$env:DATABASE_URL = 'postgres://UTENTE:PASSWORD_CODIFICATA@127.0.0.1:5432/karatty?sslmode=disable'
```

Codifica per URL i caratteri riservati della password. `sslmode=disable` vale qui solo per la base sullo stesso computer. Mantieni credenziali fuori da Git e schermate pubbliche.

In alternativa, Compose di sviluppo crea `termlink`, esponendo PostgreSQL solo su loopback:

```powershell
$env:TERMLINK_DB_PASSWORD = 'termlink-local-development-only'
docker compose -f deployments/compose.yml up -d db
$env:DATABASE_URL = "postgres://termlink:$env:TERMLINK_DB_PASSWORD@127.0.0.1:5432/termlink?sslmode=disable"
```

La password è solo un esempio locale. Non avviare il profilo `app`: conserva il prototipo inviti/SSH per i test.

## 3. Scegliere locale o LAN con TLS

**Solo questo computer:**

```powershell
go run ./scripts/devcert.go
./bin/server.exe --addr 127.0.0.1:8443 --tls-cert .local/tls/server.crt --tls-key .local/tls/server.key --public-signup
```

Entrambi i client usano `https://127.0.0.1:8443` e `.local/tls/ca.crt`. Mantieni aperto il terminale server.

**LAN:** trova l'IPv4 privata del server con `ipconfig` su Windows o `ip addr` su Linux. Sostituisci `192.168.1.50` in entrambi:

```powershell
go run ./scripts/devcert.go --out .local/tls-lan --ip 192.168.1.50
./bin/server.exe --addr 192.168.1.50:8443 --tls-cert .local/tls-lan/server.crt --tls-key .local/tls-lan/server.key --public-signup --public-network
```

`--public-network` permette ascolto fuori loopback, anche LAN, ma esige TLS/E2EE senza `--dev`. Tutti i client, anche sul server, usano `https://192.168.1.50:8443`. Scegli URL stabile prima della registrazione: lo stato cifrato è associato all'URL; cambiarlo può creare una nuova identità locale. Una prenotazione DHCP aiuta.

Consegna **solo `ca.crt`** su canale fidato, confrontando SHA-256 fuori da Karatty (`Get-FileHash PERCORSO/ca.crt -Algorithm SHA256` o `sha256sum PERCORSO/ca.crt`). `--ca-cert` la considera attendibile solo nel client, senza installazione globale. `server.key` rimane sul server; ogni partecipante conserva chiave SSH/stato propri. Non condividere directory di stato fra computer.

Consenti TCP 8443 nel firewall server solo dai client della sottorete privata, limitando regola a sottorete/profilo privato e mantenendo il firewall attivo. Reti ospiti possono isolare i dispositivi. Non esporre PostgreSQL o inoltrare porte del router. Da un altro Windows, `Test-NetConnection 192.168.1.50 -Port 8443` controlla raggiungibilità.

I certificati scadono in 30 giorni; la chiave privata CA non viene salvata e i file non sono sovrascritti. Per rinnovo/cambio IP, usa nuovo `--out`, riavvia e ridistribuisci/verifica la CA. Non disattivare la validazione TLS.

## 4. Registrare due persone

Alice genera una chiave dedicata in un altro terminale:

```powershell
ssh-keygen -t ed25519 -f .local/alice
./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice --username alice --public
./bin/karatty.exe chat --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice
```

Bob usa altro terminale/chiave:

```powershell
ssh-keygen -t ed25519 -f .local/bob
./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/bob --username bob --public
./bin/karatty.exe chat --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/bob
```

In LAN sostituisci `--server` e `--ca-cert` in **ogni** comando con URL privata e CA ricevuta. Ciascuno genera la chiave sul proprio dispositivo, senza copiare quella del contatto. Crea prima la directory se manca. Dal pacchetto estratto esegui `./karatty.exe` accanto a `karatty-crypto.exe`.

Conserva fuori dal dispositivo il codice mostrato una sola volta. Scegli/conferma frase locale indipendente dalla frase SSH e sblocca l'archivio a ogni apertura. Chiavi diverse separano Alice/Bob anche sulla stessa macchina. Se usi `--state-dir PERCORSO_ASSOLUTO`, ripetilo coerentemente.

## 5. Accettare, verificare, conversare

Nelle due TUI:

```text
Alice: /direct bob
Bob:   /requests
Bob:   /direct alice
Entrambi: Ctrl+L per ricaricare; Tab per selezionare il diretto.
Entrambi: /myfingerprint
       Condividono e confrontano le impronte tramite altro canale fidato.
Alice: /fingerprint bob
Alice: /verify bob IMPRONTA_DI_BOB
Bob:   /fingerprint alice
Bob:   /verify alice IMPRONTA_DI_ALICE
Alice: Ciao Bob
Bob:   Ciao Alice
```

Verifica testo leggibile sull'altro client e ACK del mittente. Chiudi entrambi con Ctrl+C, riaprili con medesime chiavi, URL/frasi e controlla la cronologia. ACK significa persistenza, non lettura. `/pending` e `/retry ID` ispezionano/ritentano gli invii senza conferma con lo stesso involucro.

Per validare LAN ripeti su **due macchine fisiche**. Due processi sul medesimo host/IP privato non verificano firewall remoto o isolamento Wi-Fi. Registra sistemi, terminali e risultato, senza chiavi, codici o conversazioni private. Vedi [prove esistenti](verification.md) e [recupero/eliminazione](install.md).

## Fermare e conservare dati

Ctrl+C ferma client/server; con Docker, `docker compose -f deployments/compose.yml stop db` ferma PostgreSQL. Conserva base, chiavi e stato cifrato per riutilizzare gli account. Evita `down -v` se vuoi mantenere la base.

E2EE nasconde messaggi a 720 ore e li elimina all'avvio/ogni ora. Anche la deduplicazione dura 30 giorni. La pulizia locale avviene usando il client, senza cancellare dispositivi spenti/copie esterne. Nessun backup automatico in questa guida: l'amministratore deve dichiararne/verificarne la conservazione. Gli [strumenti](operations.md) si configurano separatamente.

## Controllare lo sviluppo

```powershell
$env:TERMLINK_TEST_DATABASE_URL = $env:DATABASE_URL
$env:KARATTY_CRYPTO_BIN = (Resolve-Path bin/karatty-crypto.exe).Path
./scripts/check.ps1
```

I test creano/eliminano schemi casuali; il ruolo deve poterli creare. Senza variabili si saltano le integrazioni corrispondenti. Su Linux/macOS esporta le stesse variabili e segui [contributi](project/CONTRIBUTING.md). Per hosting Internet usa il [piano futuro](deploy.md).

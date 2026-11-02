# Installare e usare Karatty alpha

[English](../install.md) · [Español](../es/install.md) · **Italiano**

L'alpha è per uso locale con server amministrato da te. Nessun hosting del progetto su Internet; dominio/VPS sono futuri. Scarica [v0.1.0-alpha.1](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1). La [guida locale/LAN](quickstart.md) avvia server e due account. LAN su due macchine fisiche non verificata.

Gli esempi usano un server sullo stesso computer. In LAN sostituisci URL e `PERCORSO_CA` in **ogni** comando con indirizzo privato e `ca.crt` ricevuto/verificato con l'amministratore. Mantieni URL costante per conservare l'identità locale. Il client non richiede PostgreSQL.

## Scaricare e verificare

I pacchetti contengono client/bridge; il server si compila secondo la guida. Non esistono pacchetti per altre architetture.

| Piattaforma | File |
| --- | --- |
| Windows x64 | `karatty-v0.1.0-alpha.1-windows-amd64.zip` |
| Linux x64 | `karatty-v0.1.0-alpha.1-linux-amd64.tar.gz` |
| macOS Apple Silicon | `karatty-v0.1.0-alpha.1-darwin-arm64.tar.gz` |

Scarica anche `SHA256SUMS` della stessa versione e confronta prima di estrarre:

```powershell
# Windows: confronta il risultato con la riga dell'archivio scaricato.
Get-FileHash .\karatty-v0.1.0-alpha.1-windows-amd64.zip -Algorithm SHA256
```

```sh
# Linux
sha256sum -c SHA256SUMS --ignore-missing
# macOS: confronta il risultato con SHA256SUMS.
shasum -a 256 karatty-v0.1.0-alpha.1-darwin-arm64.tar.gz
```

Estrai in una directory propria e conserva `karatty`/`karatty-crypto` insieme (`.exe` su Windows); puoi aggiungerla al PATH. Sono inclusi licenze, lock e commit in `BUILD_INFO`. I pacchetti alpha non sono firmati/notarizzati. Gli hash verificano integrità rispetto al manifesto, non autenticano autonomamente l'autore.

Defender ha bloccato una compilazione server. Il proprietario l'ha ripristinata e i test sono passati con protezioni attive; la classificazione non è stata dimostrata errata. In caso di blocco conserva l'avviso e verifica provenienza, senza disattivare protezione. [Verifica](verification.md) identifica build/limiti.

## Creare l'identità

Usa una chiave SSH Ed25519 dedicata. OpenSSH chiede la frase protettiva; scegli una frase distinta per l'archivio.

```sh
ssh-keygen -t ed25519 -f PERCORSO_CHIAVE
karatty register --server https://127.0.0.1:8443 --ca-cert PERCORSO_CA --key PERCORSO_CHIAVE --username TUO_PSEUDONIMO --public
```

In PowerShell usa `./karatty.exe` dalla cartella estratta. Username: lettere ASCII, numeri, trattino/underscore, fino a 64 caratteri. Nessuna email o nome reale. Il server conserva chiave pubblica e metadati account/chat; pseudonimo non significa anonimato assoluto.

Salva il codice mostrato una volta **fuori dal dispositivo**. Scegli/conferma una frase locale di almeno 12 caratteri: non lascia il dispositivo e il codice di recupero non la sostituisce. I prompt runtime sono ancora spagnoli.

## Aprire un diretto

```sh
karatty chat --server https://127.0.0.1:8443 --ca-cert PERCORSO_CA --key PERCORSO_CHIAVE
```

Inserisci eventuale frase SSH e quella locale. Nella TUI:

1. Alice richiede `/direct bob`.
2. Bob consulta `/requests` e accetta con `/direct alice`.
3. Entrambi usano `/myfingerprint` e condividono la propria impronta su altro canale fidato.
4. Alice consulta `/fingerprint bob`, confronta il valore indipendente e usa `/verify bob IMPRONTA_DI_BOB`; Bob fa lo stesso.
5. Scrivi/Invio; Tab cambia chat, PgUp/PgDn scorrono, Ctrl+L ricarica, Ctrl+C esce.

Il confronto indipendente rileva sostituzioni al primo contatto. Un'impronta copiata solo nella medesima conversazione non è indipendente. Un'identità cambiata ferma l'invio.

ACK conferma l'involucro persistito, non la lettura. Ctrl+R ritenta con identici ID/involucro. Dopo il riavvio, `/pending` elenca i salvati e `/retry ID` li ritenta; CLI: `pending`, `retry-pending --conversation ID_CHAT --id ID_INVIO`. Esc torna all'editing mantenendo l'invio salvato. Abbandonare un invio incerto non ritira una copia già accettata. `/decline UTENTE`, `/block UTENTE`, `/unblock UTENTE` gestiscono richieste/blocco.

## Cronologia, recupero, eliminazione

Conservazione: 720 ore/30 giorni. Il server nasconde scaduti e pulisce all'avvio/ogni ora. Il client pulisce all'apertura/salvataggio e filtra letture; non pulisce dispositivi spenti. La guida non configura backup automatici: l'amministratore ne dichiara/verifica la conservazione. Il limite futuro di sette giorni aggiuntivi è requisito dell'hosting, non promessa di un servizio attivo. La segretezza in avanti non protegge copie sbloccate o dispositivi compromessi.

Se perdi SSH, genera una nuova chiave:

```sh
karatty recover --server https://127.0.0.1:8443 --ca-cert PERCORSO_CA --key NUOVA_CHIAVE --username TUO_PSEUDONIMO
```

Inserisci il codice e conserva il sostituto. Apri con la nuova chiave. Se esiste il vecchio archivio, chiudi altri client e usa la stessa frase per migrare cronologia leggibile. Lo stato vecchio viene rimosso dopo il salvataggio nuovo. Gli invii senza ACK sono in `karatty uncertain`, senza ritentativo automatico. Il recupero account non ripristina archivio/frase perduti.

I contatti confrontano la nuova impronta e usano `/reverify UTENTE IMPRONTA`. Se hanno invii pendenti, `/discardpending UTENTE` li abbandona esplicitamente prima della nuova identità; controlla se erano arrivati prima di ripeterli. I messaggi mai decifrati possono risultare indisponibili.

Eliminare l'account:

```sh
karatty delete-account --server https://127.0.0.1:8443 --ca-cert PERCORSO_CA --key PERCORSO_CHIAVE --confirm-delete
```

Rimuove account, chiavi, sessioni e messaggi propri dalla base attiva e file locali dell'account nella radice configurata. Chiudi altri client per il blocco di pulizia. Copie/dispositivi altrui non si eliminano da remoto; i backup seguono scadenze proprie. Se il server elimina ma la pulizia fallisce, il CLI indica il percorso da controllare.

## Aggiornare o disinstallare

Chiudi Karatty, verifica/estrai il nuovo pacchetto e sostituisci **entrambi** gli eseguibili, conservando chiavi/stato. Un dispositivo per account.

Percorsi predefiniti: `%AppData%\karatty` su Windows; `$XDG_CONFIG_HOME/karatty` o `~/.config/karatty` su Linux; `~/Library/Application Support/karatty` su macOS. Alternativa: `--state-dir PERCORSO_ASSOLUTO`. Rimuovere eseguibili non elimina l'account server; usa prima `delete-account` se desiderato.

Vedi [architettura](architecture.md), [E2EE](adr/0003-e2ee-1to1.md) e [contributi](project/CONTRIBUTING.md).

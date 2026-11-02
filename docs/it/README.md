<div align="center">

![Marchio terminale Karatty](../assets/karatty-mark.svg)

# Karatty

[English](../../README.md) · [Español](../es/README.md) · **Italiano**

**Conversazioni cifrate. Nel tuo terminale.**

Una TUI locale con diretti 1:1 accettati, identità SSH e crittografia end-to-end.

[![v0.1.0-alpha.1](https://img.shields.io/badge/release-v0.1.0--alpha.1-4f46e5?style=flat-square)](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1) [![CI](https://github.com/karatekah/karatty/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/karatekah/karatty/actions/workflows/ci.yml) [![Licenza Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-7ae2b5?style=flat-square)](../../LICENSE) [![Pacchetti client: Windows, Linux e macOS](https://img.shields.io/badge/platforms-Windows%20%7C%20Linux%20%7C%20macOS-64748b?style=flat-square)](install.md)

[Avvio rapido](quickstart.md) · [Scarica l'alpha](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1) · [Documentazione](#documentazione) · [Roadmap](roadmap.md) · [Segnala un bug](https://github.com/karatekah/karatty/issues)

</div>

---

Esegui l'**alpha sperimentale** sul tuo computer: un server con PostgreSQL e un client locale per partecipante. Senza dominio, VPS o email.

## Funzionalità dell'alpha

- **Identità SSH.** Registrazione senza email o nome reale, con prova del possesso di una chiave Ed25519. Recupero tramite codice conservato fuori dal dispositivo.
- **Prima il consenso.** Chat dirette con accettazione, richieste in arrivo, rifiuto, blocco e conversazioni chiuse consultabili in sola lettura.
- **E2EE locale.** Olm 1:1 con Double Ratchet tramite vodozemac 0.11.1. Le impronte vanno confrontate attraverso un altro canale fidato e verificate nuovamente dopo un cambio d'identità.
- **Archivio cifrato.** Archivio locale cifrato con frase segreta, Argon2id e XChaCha20-Poly1305. Gli invii in attesa sono salvati prima della trasmissione e sopravvivono al riavvio.
- **Consegna persistente.** Involucri opachi in PostgreSQL, cronologia per sequenza, riconnessione e deduplicazione nei 30 giorni di conservazione.
- **Ciclo di vita dei dati.** Eliminazione dell'account e dei file locali, rimozione dei messaggi scaduti e strumenti di backup cifrato e ripristino.

## Provarlo

La [guida locale e LAN](quickstart.md) spiega compilazione, PostgreSQL, server TLS e prova con due account. [Installazione e primo utilizzo](install.md) tratta pacchetti, chiavi, impronte, ritentativi e recupero. La [matrice di verifica](verification.md) distingue test automatici e prove fra macchine fisiche.

Nella TUI:

```text
Alice: /direct bob
Bob:   /requests
Bob:   /direct alice
Entrambi: /myfingerprint
       Confrontano le impronte tramite un altro canale.
Alice: /verify bob IMPRONTA_DI_BOB
Bob:   /verify alice IMPRONTA_DI_ALICE
```

Dopo la verifica, scrivi normalmente. Tab cambia chat, Ctrl+L ricarica, Ctrl+C esce. `/pending` elenca invii salvati senza conferma; `/retry ID` riutilizza lo stesso involucro. Un ACK conferma la persistenza sul server, non la lettura.

## Architettura e tecnologie

| Componente | Tecnologia |
| --- | --- |
| Client e API | Go 1.27.3, monolite modulare e contratti JSON condivisi |
| Interfaccia terminale | Bubble Tea v2 e Bubbles v2 |
| E2EE locale | Rust 1.99.1, vodozemac 0.11.1 e Olm v1 |
| Archivio locale | Argon2id, XChaCha20-Poly1305, blocco e sostituzione del file |
| Trasporto | HTTPS e WSS; autorizzazione per operazione |
| Persistenza | PostgreSQL 17.11 e pgx, sequenze e ACK transazionali |
| Hosting futuro | Linux/systemd, TLS, backup `pg_dump` cifrati con `age` |

[Architettura](architecture.md) descrive pacchetti e flussi. [ADR 0003](adr/0003-e2ee-1to1.md) documenta identità, ratchet, archivio e minacce. Il modulo Go attuale è `github.com/karatekah/karatty`. La release storica v0.1.0-alpha.1 usa il precedente nome interno del repository/modulo. Per compatibilità restano identificatori come `cmd/termlink` e le variabili `TERMLINK_*`.

## Documentazione

| Esigenza | Guida |
| --- | --- |
| Installare, conversare, aggiornare, eliminare | [Primo utilizzo](install.md) |
| Esecuzione e sviluppo locale | [Ambiente locale](quickstart.md) · [Contributi](project/CONTRIBUTING.md) |
| Capire il sistema | [Architettura](architecture.md) · [Protocollo](protocol.md) · [Dipendenze](dependencies.md) |
| Valutare la sicurezza | [Sicurezza e privacy](security.md) · [Verifica](verification.md) · [Segnalazioni private](project/SECURITY.md) |
| Futuro hosting Internet | [Piano di distribuzione](deploy.md) · [Operazioni e backup](operations.md) |
| Release e sviluppi futuri | [Pubblicazione](release.md) · [Note di versione](releases/v0.1.0-alpha.1.md) · [Changelog](project/CHANGELOG.md) · [Roadmap](roadmap.md) |
| Bozze per il futuro servizio ospitato | [Privacy](privacy-draft.md) · [Condizioni](terms-draft.md), non operative |
| Mantenere le traduzioni | [Lingue della documentazione](translations.md) |

## Sviluppo

<details>
<summary>Compilare e testare</summary>

```sh
go mod verify
cargo test --locked --manifest-path crypto/Cargo.toml
cargo build --locked --release --manifest-path crypto/Cargo.toml
go vet ./...
go test -count=1 ./...
go build ./cmd/...
```

I test d'integrazione richiedono `TERMLINK_TEST_DATABASE_URL` verso una base di sviluppo e `KARATTY_CRYPTO_BIN` verso il bridge compilato. Se mancano le variabili, i test corrispondenti vengono saltati: un risultato locale verde non sostituisce CI. Actions usa PostgreSQL reale, controllo delle race, Docker, Go/Rust, estrazione dei pacchetti su tre sistemi e controlli delle vulnerabilità.

</details>

---

Il codice del progetto usa [Apache-2.0](../../LICENSE). I pacchetti conservano licenze originali e versioni fissate delle dipendenze.

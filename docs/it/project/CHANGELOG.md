# Changelog

[English](../../../CHANGELOG.md) · [Español](../../es/project/CHANGELOG.md) · **Italiano**

## Non pubblicato — aggiornamenti di sicurezza

- Go 1.27.4 è richiesto nel modulo, CI, Docker e packaging per correggere GO-2026-6859.
- Il bridge Olm usa vodozemac 0.11.2, con correzioni per azzeramento dei segreti in memoria e chiavi Ed25519 malformate.
- I binari pubblicati di v0.1.0-alpha.1 mantengono le dipendenze originali. Le correzioni sono nel codice attuale e nei nuovi pacchetti CI.

## v0.1.0-alpha.1 — alpha sperimentale

- TUI locale per chat dirette 1:1 con accettazione, rifiuto e blocco.
- Identità pseudonima con chiave SSH Ed25519 senza email; recupero tramite codice conservato fuori dal dispositivo.
- Olm con vodozemac 0.11.1, impronte confrontate su altro canale e nuova verifica dopo un cambio d'identità.
- Archivio Argon2id/XChaCha20-Poly1305, coda persistita prima dell'invio e migrazione dopo il recupero.
- Involucri opachi in PostgreSQL, cronologia per sequenza e deduplicazione nei 30 giorni di conservazione.
- Eliminazione dell'account/stato locale, pulizia dei messaggi scaduti e strumenti di backup cifrato/ripristino.
- Pacchetti nativi client/bridge, hash, commit di origine e licenze di terzi.
- Uso locale con server TLS proprio, certificati con SAN IP e guida per due account. Dominio, VPS, caselle e hosting restano futuri.
- Prova Windows con server e due TUI: registrazione, accettazione, impronte, messaggi E2EE e riapertura della cronologia riusciti dopo il ripristino del server dalla quarantena con Defender attivo. LAN tra due macchine e rilevazione su altre installazioni non verificate.

## Migrazione del repository e della documentazione

Repository e modulo Go attuale sono `karatekah/karatty`. L'inglese è primario, con pagine complete spagnole/italiane e selezione della lingua per pagina. Tag, binari e checksum di v0.1.0-alpha.1 non cambiano; lo snapshot storico mantiene il precedente identificatore del modulo. Percorsi dei comandi e variabili compatibili conservano la grafia documentata.

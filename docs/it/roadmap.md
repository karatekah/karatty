# Ambito e decisioni di prodotto

[English](../roadmap.md) · [Español](../es/roadmap.md) · **Italiano**

L'alpha è software locale: server PostgreSQL proprio, TUI locali e diretti 1:1 accettati. Non richiede dominio, VPS o email. Hosting Internet del progetto è futuro; `karatty.dev` è scelto, non acquistato.

## Alpha locale e LAN sperimentale

- Identità Ed25519 SSH pseudonima, senza email/nome reale; codice di recupero fuori dal dispositivo.
- E2EE con segretezza in avanti circoscritta secondo [ADR 0003](adr/0003-e2ee-1to1.md), confronto esterno delle impronte e nuova verifica ai cambiamenti.
- Archivio con frase locale, coda durevole e un dispositivo per account.
- 720 ore: nascondere/eliminare corpi e ID scaduti, deduplicazione solo per 30 giorni. `--public-signup` attiva la pulizia. Copie esterne/dispositivi spenti hanno limiti separati.
- Codice, documentazione e client GitHub sotto Apache-2.0; [pubblicazione](release.md) descrive processo/limiti.
- [Verifica](verification.md) include due TUI Windows dopo il ripristino del server da Defender. LAN fisica e rilevazione su altri sistemi ancora non verificate.

## Hosting futuro

Dominio, VPS, bucket, caselle, politiche con operatore identificato e test remoti appartengono a questa fase, non sono prerequisiti del software locale. Completare [operazioni](operations.md)/[distribuzione](deploy.md) prima di offrire hosting. Finestra backup scelta: massimo sette giorni aggiuntivi, da provare prima di prometterli.

## Dopo l'alpha

Gruppi, SSH remoto con modello crittografico adeguato, dispositivi multipli, allegati, ricerca, notifiche e integrazioni necessitano progetti/test propri, senza date promesse. Community/Pro/Teams restano ipotesi.

Karatty è il nome di lavoro scelto. La ricerca preliminare ha rilevato [Karatty VR](https://usb50-zx-spectrum.itch.io/ka). RDAP indicava non registrato il 31 ottobre 2026, senza prenotare o garantire acquisto. [Indecopi](https://pi.indecopi.gob.pe/buscatumarca/#/inicio) mostrava zero corrispondenze precise attive in classi 9/38/42 il 1 novembre 2026; ricerca ampia: 200. Rimane controllo delle somiglianze.

Repository e modulo attuali usano `karatty`; il sorgente storico alpha.1 conserva `termlink`. Percorsi, variabili e domini di autenticazione compatibili mantengono la grafia. La migrazione non sostituisce binari pubblicati o tag alpha.1.

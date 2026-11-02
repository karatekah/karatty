# Scope and product decisions

**English** · [Español](es/roadmap.md) · [Italiano](it/roadmap.md)

The first alpha is local software: your PostgreSQL server, local TUI clients and accepted 1:1 direct chats. No purchased domain, VPS or email is required. A project-hosted Internet service is future work; `karatty.dev` is selected but not purchased.

## Local alpha and experimental LAN

- Pseudonymous Ed25519 SSH-key identity without email/real name; recovery through an offline code.
- E2EE with bounded forward secrecy under [ADR 0003](adr/0003-e2ee-1to1.md), out-of-band fingerprints and re-verification after identity change.
- Passphrase-protected archive, durable pending sends and one device per account.
- 720-hour retention: hide/purge expired bodies/identifiers; deduplicate only within 30 days. `--public-signup` enables purging automatically. External copies/offline devices have separate limits.
- Source, documentation and client packages on GitHub under Apache-2.0. [Release](release.md) describes publication and boundaries.
- [Verification](verification.md) includes two native Windows TUIs after the server was restored from Defender quarantine. Two physical LAN hosts and the detection on other installations remain unverified.

## Future hosting

Domain/VPS/bucket/mailbox purchases, operator-identified policies and remote backup/restore belong to that milestone, not prerequisites for distributing local software. Complete [operations](operations.md)/[deployment](deploy.md) before offering hosting. The selected maximum backup window is seven extra days and must be verified against actual infrastructure before promising it.

## After the alpha

Groups, remote SSH with an appropriate encryption model, multiple devices, attachments, search, notifications and integrations need their own design/tests. No date is committed. Community/Pro/Teams plans remain hypotheses.

Karatty is the chosen working name. Preliminary research noted the game [Karatty VR](https://usb50-zx-spectrum.itch.io/ka). RDAP returned unregistered on October 31, 2026; that neither reserves the domain nor guarantees purchase. [Indecopi](https://pi.indecopi.gob.pe/buscatumarca/#/inicio) showed zero current precise-name matches for Karatty in classes 9/38/42 on November 1, 2026; broad search returned 200 results. Similarity review remains pending.

The repository/current Go module now use `karatty`. The historical alpha.1 source retains `termlink`; compatibility build paths, environment names and authentication domains keep their existing spelling. This migration does not replace published binaries or move the alpha.1 tag.

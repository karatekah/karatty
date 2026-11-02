# Contribuire a Karatty

[English](../../../CONTRIBUTING.md) · [Español](../../es/project/CONTRIBUTING.md) · **Italiano**

Questa guida spiega come riprodurre l'ambiente dell'alpha sperimentale e proporre modifiche. Segnala problemi funzionali negli [issue](https://github.com/karatekah/karatty/issues); le vulnerabilità seguono il canale privato di [SECURITY.md](SECURITY.md).

## Preparazione

Usa Go 1.27.3, Rust 1.99.1 e PostgreSQL 17.11. La [guida locale](../quickstart.md) copre database, certificati di sviluppo e client. I test richiedono `TERMLINK_TEST_DATABASE_URL` verso una **base di sviluppo** con permesso di creare schemi e `KARATTY_CRYPTO_BIN` verso il bridge per le prove E2EE. Non usare produzione.

```sh
go mod verify
cargo test --locked --manifest-path crypto/Cargo.toml
cargo build --locked --release --manifest-path crypto/Cargo.toml
# Imposta KARATTY_CRYPTO_BIN al percorso assoluto di crypto/target/release/karatty-crypto.
go vet ./...
go test -count=1 ./...
go build ./cmd/...
```

In PowerShell puoi eseguire `./scripts/check.ps1`; per la documentazione, `./scripts/check-docs.ps1`. CI controlla anche race su Linux, build native Windows/Linux/macOS e avvio Docker. Senza la variabile della base, le integrazioni vengono saltate e non risultano validate.

## Proporre modifiche

Apri un issue prima di cambiare comportamento o protocollo. Mantieni modifiche mirate, descrivendo problema, file coinvolti, verifiche e limiti. Usa commit convenzionali: `feat:`, `fix:`, `docs:`, `test:`, `chore:`. Le migrazioni devono conservare i dati o dichiarare eventuali perdite inevitabili.

E2EE, identità, conservazione ed eliminazione riguardano sicurezza e privacy. Documenta modello delle minacce, compatibilità e prove dei guasti. Non introdurre primitive crittografiche proprie né garanzie prive di riscontro.

L'inglese è la lingua primaria. Aggiorna nella stessa modifica le versioni complete spagnola e italiana, mantenendo letterali comandi/API e seguendo la [guida alle traduzioni](../translations.md).

## Sicurezza

Non pubblicare vulnerabilità sfruttabili o dati sensibili negli issue. Usa **Security → Report a vulnerability** su GitHub e [SECURITY.md](SECURITY.md). Le segnalazioni pubbliche devono usare dati di prova, senza chiavi, token o conversazioni private.

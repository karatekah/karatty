# Bridge Olm locale

[English](../../../crypto/README.md) · [Español](../../es/crypto/README.md) · **Italiano**

`karatty-crypto` fa parte del client locale dell'alpha. Usa vodozemac 0.11.2 per identità Olm, reintegro delle chiavi monouso e cifratura/decifratura. Riceve un oggetto JSON su stdin e ne restituisce uno su stdout. I test Rust coprono scambio bidirezionale, manomissione, replay, messaggi fuori ordine e segretezza in avanti circoscritta. `pkg/cryptobridge` verifica invocazione Go e firme delle chiavi iniziali/reintegrate.

I campi `account` e `session` contengono materiale privato non cifrato mentre attraversano **pipe locali tra processi**. Non registrarli, salvarli in chiaro o inviarli al server. Il bridge non verifica l'identità del contatto: `pkg/secureclient` controlla firme/impronte, arresta l'invio ai cambi d'identità e persiste atomicamente ratchet avanzato e involucro pendente prima della trasmissione. `pkg/localstate` protegge l'archivio con frase locale, Argon2id e XChaCha20-Poly1305.

Da `crypto/`, esegui `cargo test --locked`, poi `cargo build --locked --release`; conserva il binario accanto a `karatty`. Vedi [ADR 0003](../adr/0003-e2ee-1to1.md) e [verifica](../verification.md) per prove/limiti dell'integrazione. Karatty non ha un audit indipendente.

# ADR 0002 — Inviti e prova di possesso

[English](../../adr/0002-authentication.md) · [Español](../../es/adr/0002-authentication.md) · **Italiano**

Accettato il 2026-10-31. Decisione storica del prototipo a inviti.

Registrazione/login firmano una sfida casuale monouso di breve durata, con dominio TERMLINK e chiave pubblica associata in PostgreSQL. Quel dominio di firma mantiene il nome per compatibilità. Si verifica la firma SSH prima di consumare l'invito e registrare la chiave. L'invito è rappresentato da hash, scadenza e usi residui, consumato sotto blocco nella stessa transazione. Nessuna chiave privata raggiunge il server.

SSH prova il possesso durante l'handshake e accetta solo chiavi già associate. Il prototipo consente registrazione a inviti e accesso SSH senza client installato; gli inviti possono assegnare un canale iniziale. Il flusso remoto è escluso dall'alpha. L'amministrazione usa accesso locale alla base, senza endpoint pubblico.

Token casuali rimangono nella memoria client, con hash sul server; non esistono file di token. Le chiavi cifrate si sbloccano con input nascosto e dipendono dai permessi del filesystem. In seguito sono stati aggiunti registrazione pseudonima senza email/invito e recupero autonomo con codice monouso: vedi [architettura](../architecture.md) e [protocollo](../protocol.md). Archivi di credenziali del sistema, dispositivi multipli e rotazione assistita restano futuri.

# ADR 0001 — Nucleo condiviso e PostgreSQL

[English](../../adr/0001-foundation.md) · [Español](../../es/adr/0001-foundation.md) · **Italiano**

Accettato il 2026-10-31. Registra la decisione iniziale; le tabelle di identità/chiavi E2EE sono state aggiunte con [ADR 0003](0003-e2ee-1to1.md).

Un modulo Go produce API, gateway SSH e client. SSH verifica il possesso della chiave e crea una sessione nello stesso nucleo di dominio. API/SSH autorizzano ogni operazione; la TUI condivide componenti tramite un'interfaccia piccola.

PostgreSQL conserva utenti, chiavi pubbliche, sessioni, inviti, sfide, conversazioni, appartenenze e messaggi. Nel primo sprint non furono aggiunte tabelle inutilizzate per dispositivi o ricevute di lettura. L'ACK di persistenza usa risposta/ID del messaggio; ricevute e dispositivi multipli restano futuri.

Il blocco della conversazione assegna sequenze transazionali. L'unicità conversazione/mittente/ID deduplica nei 30 giorni di conservazione; contenuto diverso con lo stesso ID genera conflitto. La scadenza elimina anche l'ID; dopo, l'invio è nuovo. Non ci si affida all'orologio client.

Il tempo reale iniziale legge il registro ogni 250 ms via WSS o adattatore del nucleo, ricontrollando sessione, chiave, utente e appartenenza per lotto. Evita la corsa fra cronologia/sottoscrizione, supporta più processi e non richiede Redis. Costo: query periodiche e fino a 250 ms di latenza; misurare prima di aggiungere LISTEN/NOTIFY.

Le operazioni mantengono blocchi condivisi d'autorizzazione durante la transazione; la revoca attende quelle in corso e i lotti successivi falliscono. Chiavi/sessioni revocate chiudono la sottoscrizione al prossimo controllo. Il gateway rivaluta anche SSH inattivo.

Pacchetti users/messages/realtime separati verranno creati quando avranno responsabilità autonome. Nel primo sprint le operazioni appartengono ad auth/chat/transport, senza directory vuote convenzionali.

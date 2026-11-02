# Sicurezza e privacy

[English](../security.md) · [Español](../es/security.md) · **Italiano**

I test richiesti coprono isolamento dei membri, revoca, inviti, firma/riuso delle sfide, sequenze, deduplicazione e persistenza. Non registrare corpi, token, inviti, firme o chiavi private. I log operativi contengono solo avvio/errori generici e contatori aggregati.

## Trasporto e limiti

Messaggi leggibili del prototipo: massimo 4096 byte UTF-8, senza controlli o formattazione invisibile; involucri E2EE: 12288 byte. Pagine di 100 messaggi; connessioni, handshake, frequenze e tempi di scrittura sono limitati. I consumatori lenti si disconnettono e sincronizzano dal registro durevole; la TUI limita la cronologia visualizzata.

TLS è obbligatorio salvo HTTP di sviluppo esplicitamente abilitato solo su loopback. Il prototipo SSH usa una chiave host persistente da verificare; non disattivare `StrictHostKeyChecking`. L'ascolto predefinito è loopback.

La [matrice di verifica](verification.md) copre E2EE, perdita/cambio d'identità, archivio, limiti e pacchetti. Documenta la rilevazione Defender e la prova Windows riuscita dopo il ripristino del file da parte dell'utente, con protezioni attive. Non conferma un falso positivo o altre installazioni. LAN fra due macchine non verificata; nessun audit indipendente di Karatty.

La registrazione aperta usa pseudonimo e chiave SSH senza email. La TUI cifra con Olm dopo il confronto delle impronte. Inviti/SSH remoto non hanno E2EE e sono esclusi dall'alpha. L'hosting futuro richiede controlli di provider, backup e politiche.

## Stato locale e recupero

Argon2id e XChaCha20-Poly1305 cifrano l'archivio con una frase richiesta all'apertura e mai inviata al server. Le voci scadute vengono rimosse all'apertura/salvataggio; un dispositivo spento non può eseguire la pulizia. Dopo il recupero SSH, la cronologia leggibile viene migrata in un nuovo file cifrato e il vecchio stato eliminato solo dopo salvataggio/sincronizzazione. Gli invii precedenti senza ACK rimangono incerti per revisione manuale. Non è garantita pulizia quando il programma non gira.

Riaprire un diretto richiede entrambe le appartenenze e non ripristina accessi revocati. Il recupero usa codice offline e nuova chiave SSH. Il client cifrato salva involucro e ID prima di trasmettere; il ritentativo li conserva esattamente. Il client precedente conserva i pendenti solo in memoria: gli script devono riutilizzare `--id`.

## Limiti per processo

| Risorsa | Limite |
| --- | --- |
| API | 10 richieste/s per IP, raffica 30 |
| Invii | 2 messaggi/s per utente, raffica 10 |
| SSH del prototipo | 1 tentativo d'autenticazione/s per IP, raffica 10 |
| Identità del limitatore | 4096 |
| Connessioni HTTP / SSH | 128 / 64 |
| Sottoscrizioni/canali SSH | 64 per processo |
| Sfide di autenticazione | Massimo 4096, TTL di due minuti |
| Heartbeat WSS | Ogni 15 s; timeout scrittura/heartbeat 5 s |
| Handshake SSH / sessione massima | 10 s / 8 h |

I limiti sono locali al processo. Valuta applicazione condivisa e capacità prima di aggiungere processi o esporre su Internet.

## Consenso e anonimato

Un diretto richiede la chiamata reciproca del destinatario. Le richieste scadono in 7 giorni, massimo 10 uscenti attive per utente, frequenza 1/minuto e raffica 3 per processo. Si possono rifiutare o bloccare. Il blocco chiude il diretto e impedisce nuove richieste/messaggi, senza cancellare copie già consegnate.

I codici di recupero hanno 192 bit casuali, vengono conservati come hash e ruotati all'uso. La sfida dura due minuti e consente un tentativo. Il recupero revoca chiavi/sessioni precedenti. Perdendo chiave e codice non esiste recupero amministrativo via email. Pseudonimi, chiavi pubbliche, relazioni e metadati possono identificare persone: pseudonimia non significa anonimato assoluto.

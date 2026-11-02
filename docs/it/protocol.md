# Protocollo v1

[English](../protocol.md) · [Español](../es/protocol.md) · **Italiano**

JSON su HTTPS. Tutti gli endpoint salvo sfida/prova d'autenticazione, recupero e health richiedono `Authorization: Bearer SESSION_TOKEN`; niente token negli URL. Risposte `Cache-Control: no-store`. Errori: 400 input invalido, 401 autenticazione, 403 appartenenza, 409 conflitto ID/contenuto, 429 frequenza. Gli errori interni non espongono query o dati.

L'autenticazione restituisce `public_mode`. Se true, il client richiede messaggistica cifrata; `--legacy` è limitato al prototipo a inviti per sviluppo locale.

| Operazione | Percorso |
| --- | --- |
| Sfida registrazione/login | POST /v1/auth/challenge |
| Firma SSH serializzata ([]byte JSON in base64) | POST /v1/auth/proof |
| Avvio recupero con pseudonimo/nuova chiave | POST /v1/recovery/start |
| Firma nuova chiave/codice di recupero | POST /v1/recovery/complete |
| Chiusura/revoca sessione | DELETE /v1/session |
| Eliminazione account autenticato | DELETE /v1/account |
| Conversazioni accessibili, massimo 100 | GET /v1/conversations |
| Richiedere diretto o accettare richiesta | POST /v1/direct |
| Richieste dirette in arrivo | GET /v1/direct/requests |
| Rifiutare richiesta | DELETE /v1/direct/requests/{username} |
| Bloccare utente/chiudere diretto | PUT /v1/direct/blocks/{username} |
| Sbloccare utente | DELETE /v1/direct/blocks/{username} |
| Caricare identità Olm/chiavi monouso firmate | POST /v1/e2ee/device |
| Leggere identità dopo l'accettazione | GET /v1/e2ee/peer/{username} |
| Consumare chiave monouso del contatto | POST /v1/e2ee/peer/{username}/claim |
| Cronologia crescente per sequenza | GET /v1/conversations/{id}/messages?after=0&limit=100 |
| Inviare client_id stabile/body | POST /v1/conversations/{id}/messages |
| Eventi per sequenza con intestazione Bearer | WSS /v1/conversations/{id}/events?after=0 |

## Persistenza e cronologia

`protocol.Message` è anche l'ACK: ID server, `client_id`, conversazione, mittente, username, sequenza, body e data, restituiti dopo il commit. Non certifica lettura o consegna a tutte le sessioni. Mittenti diversi possono usare lo stesso ID; un mittente riutilizza esattamente l'ID per ritentare quel contenuto nella stessa conversazione.

La deduplicazione dura i 30 giorni di conservazione. La pulizia elimina contenuti e ID; un successivo invio con quell'ID è nuovo. Con `--public-signup` la pulizia è automatica all'avvio/ogni ora; in sviluppo è manuale o abilitata con `server --retention-30d`.

`Page` contiene `messages` e `next`. Il vecchio client tiene in memoria l'ultima sequenza e filtra i duplicati. Il client sicuro persiste ratchet, archivio e cursore server: riaprire la TUI ricostruisce la vista dall'archivio, sincronizzando col cursore durevole. La sequenza è assegnata sotto blocco nella transazione INSERT. Pulizia/eliminazione possono lasciare buchi; `next_seq` non arretra. La cronologia nasconde messaggi oltre 720 ore prima della rimozione fisica. Anche lotti vuoti richiedono autorizzazione. Massimo 100 per query; consumatori lenti esauriscono il timeout e si riconnettono al registro.

`Conversation.closed` indica il diretto archiviato dopo l'eliminazione di un partecipante. Il membro rimasto può leggere cronologia autorizzata conservata, non inviare. L'eliminazione rimuove chiavi, sessioni, appartenenze e messaggi dell'account, non i messaggi altrui o copie già ricevute. I backup hanno una politica separata.

## Distribuzione delle chiavi E2EE

Gli endpoint sperimentali seguono [ADR 0003](adr/0003-e2ee-1to1.md). `POST /device` esige firma SSH su `KARATTY-E2EE-IDENTITY-v1\nUSER_ID\nED25519\nCURVE25519`. Ogni chiave monouso esige firma Olm Ed25519 su `KARATTY-OTK-v1\nCURVE25519\nONE_TIME_KEY`. Base64 standard senza padding.

L'identità non si sostituisce silenziosamente; il recupero elimina quella precedente. Solo membri di un diretto accettato consultano/consumano chiavi del contatto. Una chiave consumata non si riattiva caricandola nuovamente. Il client confronta impronte su altro canale, cifra con Olm e conserva ratchet nell'archivio cifrato.

## Registrazione aperta e involucri

`server --public-signup` disattiva gli inviti e registra pseudonimi/chiavi Ed25519 senza email. Elenca solo diretti e rifiuta cronologia, eventi e invii di canali anche preesistenti. `body` deve essere un involucro JSON con `version:1`, `session_id`, `sender_identity`, `message:{type,body}`; chiavi/corpo Olm sono base64 standard senza padding. L'API verifica la **struttura**, non la cifratura corretta di client modificati. I client ufficiali elaborano gli involucri localmente.

`auth/challenge` usa `KARATTY-PUBLIC-REGISTER-v1`; `auth/proof` consuma la sfida in un tentativo e richiede firma della nuova chiave. La risposta contiene sessione temporanea e codice di 192 bit mostrato una volta, anche nella registrazione a inviti.

`recovery/start` riceve pseudonimo/nuova chiave senza rivelare se l'account esiste. `recovery/complete` richiede firma e codice, revoca chiavi/sessioni precedenti e ruota il codice. Le sfide scadono dopo due minuti.

## SSH del prototipo e ascolto

SSH verifica il possesso della chiave e crea una sessione di dominio per canale. Una shell PTY apre una TUI sul server. Comandi: `list`, `requests`, `direct USERNAME`, `decline USERNAME`, `block USERNAME`, `unblock USERNAME`, `history CONVERSATION AFTER`, `send CONVERSATION CLIENT_ID` con body su stdin. Nessun comando di sistema. **Non è E2EE ed è escluso dall'alpha.** `--public-signup` insieme a `--ssh-host-key` è rifiutato.

L'ascolto aperto usa loopback; `--public-network` richiede TLS e rifiuta `--dev`/TUI SSH remota. Vedi [guida locale/LAN](quickstart.md) e [hosting futuro](deploy.md). Infrastruttura e servizio ospitato restano futuri.

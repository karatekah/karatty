# Verifica dell'alpha sperimentale

[English](../verification.md) · [Español](../es/verification.md) · **Italiano**

Prove riproducibili del codice locale/LAN, non audit indipendente o prova di rete fra macchine fisiche. VPS/dominio/provider appartengono all'hosting futuro. Risultati per commit in [GitHub Actions](https://github.com/karatekah/karatty/actions/workflows/ci.yml).

| Area | Evidenza automatica |
| --- | --- |
| Registrazione/recupero senza email | `TestAnonymousSignupAndOfflineRecovery`: possesso, sfida monouso, codice errato, revoca/rotazione |
| Consenso/accesso alle chiavi | `TestE2EEKeyDistributionRequiresAcceptedDirect`: niente identità/consumo prima dell'accettazione; firme errate rifiutate; chiavi consumate non riattivabili |
| Messaggi E2EE | `TestE2EEClientExchangeAndLocalPassphrase`: impronte coincidenti, invio/risposta reali, PostgreSQL opaco, 4096 byte, riavvio, ACK alterato, involucro manomesso e ritentativo stabile |
| Recupero dispositivo | Sostituzione SSH/Olm, nuova verifica, migrazione archivio, invii incerti conservati, vecchio stato eliminato |
| Chiusura/eliminazione | Archivio del diretto chiuso leggibile, nuovi invii rifiutati, file account rimossi; suite server coprono autorizzazione/eliminazione |
| Segretezza in avanti circoscritta | Rust avanza il ratchet: copie attuali account/sessione non decifrano un messaggio consumato. Replay/manomissione rifiutati, fuori ordine supportato. Esclusi archivi sbloccati/vecchi snapshot |
| Archivio | `pkg/localstate`: frase errata, alterazione, blocco, sostituzione/riapertura 16 MiB. `pkg/secureclient`: 720 ore e ordine quando ACK precede arretrati |
| Limiti trasporto | `TestMaximumEnvelopePageOverHTTPAndWebSocket`: 100 involucri massimi con doppio escape JSON senza troncamento |
| TUI | `pkg/tui`: impronte intere, confronto esterno e ritentativo stesso ID |
| Client nativo | `TestE2EEClientNativeTUI`: due processi PTY reali sbloccano archivi, confrontano/verificano e consegnano messaggio memorizzato opaco |
| TLS locale/LAN | `scripts/TestDevelopmentCertificateLAN`: catena, localhost, loopback IPv4/IPv6, SAN IP privata; rifiuto IP assente, sovrascrittura, indirizzi invalidi/indefiniti/multicast |
| Distribuzione | CI compila/estrae/esegue versione/bridge sui tre sistemi, controlla hash/licenze/commit |
| Servizio | PostgreSQL reale, TLS/WSS, client eseguibile e gateway legacy; race, `go vet`, `govulncheck`, Cargo.lock/RustSec, Docker, shell e systemd |

## Prova Windows del 1 novembre 2026

`TestE2EEClientExchangeAndLocalPassphrase` è passato con PostgreSQL 17.11 nativo e bridge Windows del candidato `d39b5da`. Comprende due client su TLS loopback, scambio, cronologia, recupero e ritentativi. Quel test non è una sessione manuale con due TUI Windows. Il test PTY automatico sopra gira su Linux in CI.

Sono passati `go vet ./...`, `go test -count=1 ./...` con base/bridge, `go build ./cmd/...`, certificati e link Markdown relativi. La compilazione non dimostra che l'antivirus consentirà l'esecuzione.

Nel primo tentativo Defender ha bloccato `.local/lan-alpha/server.exe` su IP privata come `Behavior:Win32/DefenseEvasion.A!ml`. Nessuna protezione disattivata o esclusione aggiunta. Falso positivo non dimostrato; quella prova dell'ascolto IP non è riuscita.

### Prova nativa dopo il ripristino

L'utente ha ripristinato `.local/defender-review/server.exe` dalla quarantena. La build `4c0356f` ha SHA-256 `9a81de63cc5149d551bcc4abc3cf0b6a8b83b8658a58d6fc819bb8fe7db830f1`. Sono riusciti `--help` e avvio nativo su `https://127.0.0.1:28443` in modalità E2EE senza gateway SSH.

Estratto il client Windows `v0.1.0-alpha.1` dello stesso commit, sono state aperte due TUI ConPTY sullo stesso computer: registrazione con chiavi/frasi distinte, richiesta/accettazione, condivisione delle impronte proprie e verifica reciproca. Alice ha inviato, Bob risposto; entrambi i messaggi sono apparsi sull'altro client con ACK 1/2. Chiudendo/riaprendo entrambe, la cronologia ha recuperato due messaggi e sincronizzato alla sequenza 2.

PostgreSQL conteneva due involucri E2EE senza testi leggibili di prova; nemmeno i log operativi li contenevano. Client e server sono stati fermati alla fine.

Protezioni in tempo reale/comportamento attive con firme `1.463.83.0`; nessuna nuova rilevazione osservata per quel binario durante il test. Valida la build ripristinata su quel dispositivo, non un falso positivo o altre installazioni. Nessun campione inviato a Microsoft.

Era disponibile un solo computer: **LAN fra due macchine fisiche non provata**. [Istruzioni](quickstart.md) e certificati non dimostrano raggiungibilità fra macchine.

## Limiti delle prove

- Nessun audit esterno di Karatty. L'audit vodozemac riguarda la libreria/versione esaminata, non automaticamente integrazione/versioni successive.
- Testati salvataggio prima d'invio/visualizzazione, blocchi e sostituzione; non interruzioni fisiche sui tre filesystem.
- Sequenza, data, nome account e disponibilità dipendono dal server. Il cifrato lega conversazione/mittente/ID; il server può nascondere messaggi o negare servizio.
- Archivio limitato a 16 MiB prima della cifratura. Se pieno, il salvataggio fallisce prima di trasmettere ratchet non persistito. Pulizia solo col client attivo, senza cancellare dispositivi spenti/snapshot altrui.
- Mancano revisione completa con lettori di schermo e verifica visiva di tutti i terminali. CI controlla comportamento/binari.
- Hosting reale, scadenza/ripristino remoto, dominio/certificato e politiche/contatto dell'operatore restano futuri. Registrarli in [pubblicazione](release.md) prima di annunciare hosting.

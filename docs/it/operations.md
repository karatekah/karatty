# Operazioni del futuro servizio ospitato

[English](../operations.md) · [Español](../es/operations.md) · **Italiano**

Dominio, VPS e provider sono rinviati, non prerequisiti dell'[alpha locale/LAN](quickstart.md). Nessuna infrastruttura acquistata o copia remota attiva. `karatty.dev` è scelto ma non acquistato/confermato dal registrar. Mancano caselle pubbliche privacy/supporto. Gli account non richiedono email.

Nome di lavoro **Karatty**. Il 1 novembre 2026, [Busca tu marca](https://pi.indecopi.gob.pe/buscatumarca/#/inicio) restituiva **zero corrispondenze precise** attive nelle classi 9/38/42; ricerca ampia: 200. È preliminare, non autorizzazione legale. Verificare somiglianze, dominio e marchio prima dell'hosting.

## Infrastruttura e budget

Tetto futuro scelto: USD 20/mese. L'opzione [DigitalOcean Basic](https://www.digitalocean.com/pricing/droplets) registrata ha 1 vCPU, 2 GiB RAM, 50 GB SSD; Go/PostgreSQL possono condividerla con capacità controllata. [R2](https://developers.cloudflare.com/r2/pricing/) ha quota gratuita Standard; `.dev` si paga annualmente. Ricontrollare prezzi/quote prima di acquisto o espansione.

**Dati registrati il 1 novembre 2026, non preventivo attuale:** DigitalOcean USD 12/mese; R2 Standard 10 GB-mese, un milione operazioni A/dieci milioni B mensili gratuiti; [Cloudflare Registrar](https://pricing.registrar.cloudflare.com/) rinnovo `.dev` USD 12.20/anno, con eventuali promozioni iniziali diverse. Stima base 12 + 12.20/12 ≈ **USD 13.02/mese**, prima di imposte/eccedenze/altri servizi. [RDAP](https://pubapi.registry.google/rdap/domain/karatty.dev) rispondeva 404: indicava assenza di registrazione allora, senza prenotare o garantire disponibilità/prezzo finale.

[Cloudflare Email Routing](https://developers.cloudflare.com/email-service/platform/pricing/) offre inoltro in entrata per possibili `support@karatty.dev`/`privacy@karatty.dev`. [Non invia/risponde dal dominio](https://developers.cloudflare.com/email-service/reference/postmaster/#sending-or-replying-to-an-email-from-your-cloudflare-domain): la risposta parte dalla casella destinataria. L'operatore deve scegliere/verificare indirizzi prima delle politiche. Sono proposte, non contatti attivi.

## Backup

Bucket R2 dedicato privato con credenziali rclone limitate; `BACKUP_REMOTE=r2:NOME_BUCKET/db-backups`. Genera identità age altrove, mantieni chiave privata fuori dal VPS e configura solo destinatario pubblico `age1…` in `BACKUP_RECIPIENT`.

[backup.sh](../../scripts/backup.sh) passa `pg_dump -Fc` nella cifratura prima dell'upload, conserva il file cifrato locale solo durante il trasferimento ed elimina oggetti remoti più vecchi di tre giorni. Esecuzione giornaliera systemd con `Persistent=true`.

Configura anche [regola lifecycle R2](https://developers.cloudflare.com/r2/buckets/object-lifecycles/) di **tre giorni** su `db-backups/`. Verifica applicazione, versioning e altre copie. Cloudflare descrive rimozione normalmente entro 24 ore, ma potenzialmente più tardi. Per promettere massimo sette giorni aggiuntivi bisogna monitorare la rimozione reale e verificare oggetti oltre quattro giorni; il lifecycle non garantisce cancellazione fisica esatta.

Prima dell'hosting, carica/ripristina in base vuota `termlink_restore_`. Sul computer con identità privata imposta `BACKUP_OBJECT`, `AGE_IDENTITY_FILE`, `RESTORE_DATABASE_URL` ed esegui [restore-check.sh](../../scripts/restore-check.sh). Rifiuta altri nomi/tabelle esistenti. Verifica righe/accesso del servizio, ripeti periodicamente/dopo migrazioni. Mai identità privata nel bucket/VPS.

**Prova locale del 31 ottobre 2026:** backup cifrato su backend rclone locale, ripristino di tre utenti/nove messaggi in base temporanea; il backup seguente eliminava una copia marcata cinque giorni prima. Non verifica R2 o il lifecycle.

## Prima di offrire hosting

1. Confermare/acquistare dominio, verificare somiglianze e creare contatto pubblico.
2. Rivedere E2EE 1:1, escludere testo leggibile/canali/SSH remoto e provare impronte, perdita/rotazione.
3. VPS/bucket privati, DNS, TLS, amministrazione; credenziali fuori da Git.
4. Completare [privacy](privacy-draft.md)/condizioni con identità/contatto, conservazione e adempimenti peruviani applicabili.
5. Provare registrazione, recupero, consenso, blocco, eliminazione, pulizia, backup/ripristino e cancellazione sull'infrastruttura acquistata.
6. Partire da [release verificata](release.md), provando installazione da altra rete.

Registrare risultati datati prima dell'annuncio. La distribuzione locale segue una lista separata.

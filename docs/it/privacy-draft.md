# Privacy di Karatty — bozza da revisionare

[English](../privacy-draft.md) · [Español](../es/privacy-draft.md) · **Italiano**

**Bozza inattiva per hosting futuro; non presentarla come politica operativa.** L'alpha è un'istanza locale/LAN amministrata da chi la installa, senza servizio centrale. Ogni amministratore dichiara trattamento/copie della propria istanza. `karatty.dev` scelto, non acquistato. Mancano identità/contatto, provider effettivi, verifica del marchio e revisione legale/operativa. Decorrenza `[DA DEFINIRE]`.

## Titolare e contatto proposti

Futuro operatore: `[NOME LEGALE OPERATORE]`, persona fisica domiciliata in Perù. Richieste privacy/accesso/rettifica/cancellazione/opposizione: `[EMAIL PUBBLICA PRIVACY]`; supporto: `[EMAIL PUBBLICA SUPPORTO]`.

## Dati e finalità

- **Account/accesso:** pseudonimo, chiave pubblica SSH/impronta, date creazione/accesso, hash sessioni/codice di recupero. Registrazione, autenticazione, recupero, prevenzione abusi. Nessuna email account, nome reale, chiave SSH privata o codice leggibile conservato.
- **Chiavi pubbliche E2EE sperimentali:** identità Olm, firma SSH, chiavi monouso firmate e indicatori di consumo. Solo per contatti di diretti accettati e prevenzione del riuso. Il server non riceve identità privata o ratchet. Le chiavi pubbliche sono rimosse eliminando o recuperando l'account con nuovo SSH.
- **Archivio locale cifrato:** cronologia leggibile, invii pendenti e ratchet cifrati con frase indipendente sul dispositivo. Uso/apertura richiedono frase e puliscono scaduti. Perderla impedisce l'accesso; il codice non la sostituisce. Cambiando SSH si migra la cronologia leggibile ed elimina il vecchio stato dopo il salvataggio nuovo; invii non confermati rimangono incerti.
- **Messaggi:** E2EE conserva involucri opachi, mittente/conversazione/ID di ritentativo/sequenza/data, richieste e blocchi. Il prototipo a inviti/canali conserva corpi leggibili ed è escluso. I metadati servono a consegna, consenso e blocco.
- **Operazioni:** contatori aggregati e limiti temporanei in memoria. I log applicativi non contengono corpi, codici o token. I provider possono avere registri propri.

Lo pseudonimo non garantisce anonimato assoluto: chiavi, relazioni, tempi e contenuti condivisi possono identificare persone. I partecipanti leggono messaggi consegnati. Il client cifra prima d'inviare; l'operatore conserva metadati. Legacy espone testo ed è escluso. HTTPS/WSS proteggono trasporto. Nessun servizio ospitato del progetto è operativo.

## Provider e trasferimenti

Server/PostgreSQL: `[PROVIDER, REGIONE, PAESE]`. Backup cifrati: `[PROVIDER, REGIONE, PAESE]`. Codice di recupero mostrato una volta localmente, senza email. Completare luoghi/flussi internazionali prima dell'hosting.

Dichiarazione proposta, da verificare: dati personali non venduti e messaggi non usati per pubblicità/addestramento. `[VERIFICARE CONTRATTI E CONFIGURAZIONI EFFETTIVI]`.

## Conservazione ed eliminazione

- Cronologia indisponibile dopo **720 ore/30 giorni**. Pulizia fisica E2EE all'avvio/ogni ora: possibile ritardo rispetto al termine d'accesso.
- Richieste dirette: **7 giorni**; sfide: **2 minuti**; sessioni: **8 ore**. Manutenzione elimina scaduti/sessioni revocate.
- L'eliminazione rimuove chiavi SSH/E2EE, sessioni, appartenenze e messaggi scritti dalla base attiva e file locali dell'account sul dispositivo configurato. Non cancella copie/dispositivi altrui. I messaggi altrui seguono la propria scadenza; i diretti coinvolti diventano sola lettura.
- Backup completi previsti cifrati prima dell'upload e scadenti in **3 giorni**. Dati cancellati/scaduti possono permanere al massimo **7 giorni aggiuntivi** secondo la politica proposta, da verificare contro rimozione reale prima di pubblicarla. Nessun backup automatico nella guida locale.

## Scelte e diritti

Si possono rifiutare richieste, bloccare utenti, eliminare l'account. Inviare richieste d'accesso/rettifica/cancellazione/opposizione al futuro contatto privacy; potrebbe servire prova ragionevole del controllo account. Vedi [diritti ARCO](https://www.gob.pe/9270-que-son-los-derechos-arco) e reclami presso l'autorità peruviana.

## Sicurezza e modifiche

Chiave SSH e codice mostrato una volta proteggono l'accesso. Mantenerli fuori portata di terzi; perderli entrambi impedisce recupero automatico. Modifiche materiali sarebbero annunciate in `[URL PUBBLICO]` prima della decorrenza. Il recupero sostituisce SSH/Olm e conserva cronologia già leggibile solo se archivio/frase restano disponibili. I contatti devono verificare la nuova identità.

## Revisione prima dell'hosting

Rivedere indicazioni ANPD sulla [registrazione di banche dati personali](https://www.gob.pe/institucion/anpd/pages/8060-inscribir-banco-de-datos-en-el-registro-nacional-de-proteccion-de-datos-personales), inclusi titolari persone fisiche, e sui [flussi transfrontalieri](https://www.gob.pe/9253-inscribir-flujo-transfronterizo-de-datos-personales) dopo scelta dei provider. Richiedono dati del titolare; questa bozza non registra pratiche completate.

La revisione originaria cita [Decreto Supremo 016-2024-JUS](https://www.gob.pe/institucion/anpd/normas-legales/6554453-16-2024-jus), registrato come vigente dal 31 marzo 2025 secondo ANPD. L'operatore deve verificare gli obblighi applicabili prima del lancio. Consultare fonti o tradurre non costituisce validazione legale/registrazione.

# Lingue della documentazione

[English](../translations.md) · [Español](../es/translations.md) · **Italiano**

L'inglese è primario nei percorsi normali: `README.md`, politiche alla radice e guide `docs/`. Versioni complete spagnole/italiane sono in `docs/es/` e `docs/it/`. Ogni pagina porta alla medesima pagina nelle altre lingue.

## Struttura

| Sorgente inglese | Spagnolo | Italiano |
| --- | --- | --- |
| `README.md` | `docs/es/README.md` | `docs/it/README.md` |
| `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md` | Stesso nome sotto `docs/es/project/` | Stesso nome sotto `docs/it/project/` |
| `docs/install.md` e altre guide | `docs/es/install.md` e corrispondenti | `docs/it/install.md` e corrispondenti |
| `docs/adr/0003-e2ee-1to1.md` | `docs/es/adr/0003-e2ee-1to1.md` | `docs/it/adr/0003-e2ee-1to1.md` |
| `crypto/README.md`, `third_party/README.md` | Sottodirectory corrispondente in `docs/es/` | Sottodirectory corrispondente in `docs/it/` |

Il catalogo è `docs/languages.json`: comprende note, operazioni e bozze inattive, non solo README. Licenze/copyright originali rimangono autorevoli e non vengono tradotti.

## Aggiornare una pagina

1. Aggiorna inglese e entrambe le traduzioni complete nella stessa modifica, preservando qualificazioni di sicurezza, stato sperimentale, numeri, versioni e date.
2. Mantieni letterali eseguibili, flag, variabili, API, JSON e domini firmati. Traduci spiegazioni, titoli, commenti e segnaposto chiaramente esemplificativi.
3. Usa link relativi mantenendo la lingua scelta; i collegamenti a codice/configurazioni/licenze ritornano al file reale.
4. Mantieni la barra a tre lingue dopo il titolo, con lingua corrente marcata.
5. Esegui `./scripts/check-docs.ps1` in PowerShell: controlla catalogo, tre versioni, navigazione e destinazioni relative. Rivedi poi Markdown, tabelle e blocchi.

Il controllo non dimostra equivalenza semantica: serve confronto tecnico del revisore. Non sostituire traduzioni con riassunti o fallback inglesi silenziosi. I messaggi runtime della TUI non vengono tradotti da questa modifica.

## Nome e release

Repository/modulo canonici: `karatekah/karatty`. Non riutilizzare il vecchio nome del repository, preservando i redirect GitHub. I nomi di percorsi/variabili compatibili restano documentati.

Binari, checksum e tag `v0.1.0-alpha.1` non cambiano; sorgenti/documenti archiviati usano precedente modulo/lingua. Guide attuali e link linguistici della Release portano alla documentazione mantenuta, senza riscrivere lo snapshot storico.

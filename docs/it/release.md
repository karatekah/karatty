# Pubblicazione di Karatty v0.1.0-alpha.1

[English](../release.md) · [Español](../es/release.md) · **Italiano**

La [prima alpha](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1) distribuisce sorgenti, documentazione e client per uso locale con server proprio. È sperimentale. LAN documentata, non verificata fra due macchine fisiche. Nessun hosting Internet del progetto.

## Ambito ed evidenze

| Parte | Stato |
| --- | --- |
| E2EE, recupero, ritentativi, conservazione | Implementati; [prove/limiti](verification.md) |
| Windows | Server TLS/due TUI: registrazione, accettazione, impronte, scambio e riapertura |
| Defender | Necessario ripristino dalla quarantena; protezioni attive. Falso positivo/altre installazioni non determinati |
| Pacchetti | Windows x64, Linux x64, macOS Apple Silicon; client/bridge, licenze, lock, BUILD_INFO/hash. Server separato |
| LAN fisica | Pendente, configurazione sperimentale |
| Dominio/VPS/provider | Futuri, non richiesti localmente |
| Privacy/condizioni future | Bozze inattive; ogni amministratore dichiara istanza/copie |
| Vulnerabilità | Canale privato secondo [SECURITY.md](project/SECURITY.md) |

Codice proprio Apache-2.0, dipendenze con licenze proprie. Download senza firma/notarizzazione, nessun audit indipendente.

## Controlli per pubblicare

1. Registrare CI sul commit etichettato: PostgreSQL, E2EE, TUI, Go/Rust, race, estrazione e avvisi. La prova manuale Windows identifica la build precisa, non altri binari/dispositivi.
2. Controllare cronologia/file per segreti, dati personali, chiavi, dump/configurazioni. Rivedere log Actions conservati prima che diventino pubblici. Ignorare `.local/`, `.env`, `bin/`, `dist/`.
3. Rivedere README, installazione, architettura, sicurezza, contributi/changelog; mantenere visibili limiti Defender/LAN e bozze future inattive.
4. Integrare in `main`, preservare commit convenzionali, verificare CI finale. Scaricare pacchetti e controllare SHA-256, licenze e commit `BUILD_INFO`.
5. Tag `v0.1.0-alpha.1` su quel commit, tre pacchetti e `SHA256SUMS` comune, note con prove/limiti.
6. Repository pubblico, segnalazioni private abilitate, Release **prerelease** e controllo anonimo di README, tag, pagina/download.

Prima di dichiarare LAN verificata serve la prova fra macchine. Dominio, VPS, caselle e provider appartengono all'[hosting futuro](operations.md).

## Compilare i pacchetti client

Sulla piattaforma destinazione con PowerShell, Go 1.27.4, Rust 1.99.1:

```powershell
./scripts/build-release.ps1 -Version v0.1.0-alpha.1
```

Lo script compila `karatty` senza CGo e `karatty-crypto` Rust nativi. Crea directory/archivio `dist/VERSIONE/`, hash binari/archivio, licenze, lock e commit `BUILD_INFO`. Non sovrascrive output esistente. Chiavi, configurazioni private e base non sono incluse. Server secondo [guida](quickstart.md).

CI compila/estrae/esegue sui tre sistemi; gli artifact Actions sono input, non sostituti della pubblicazione/verifica. [Installazione](install.md) tratta hash, coppia di eseguibili, identità e chat.

## Documentazione dopo la release

L'inglese è primario con versioni complete [spagnola](../es/release.md)/[italiana](release.md); vedi [manutenzione](translations.md). Repository `karatekah/karatty`; binari/checksum/commit `5e4adde` e tag alpha.1 immutabili. Quel sorgente usava il vecchio modulo; lo sviluppo usa quello nuovo. Le [note](releases/v0.1.0-alpha.1.md) descrivono gli artifact pubblicati.

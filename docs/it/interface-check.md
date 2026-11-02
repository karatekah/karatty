# Revisione iniziale dell'interfaccia

[English](../interface-check.md) · [Español](../es/interface-check.md) · **Italiano**

La revisione iniziale usava `make-interfaces-feel-better` per la TUI Bubble Tea v2/Bubbles v2 condivisa: testo monospaziato, separatori e stati scritti, senza CSS o animazioni proprie. Non è una revisione completa visiva/accessibile dei terminali Windows/macOS/Linux.

| Categoria | Evidenza | Risultato |
| --- | --- | --- |
| Tipografia | `ansi.Wrap/Truncate` e larghezza terminale | Adattamento presente; controllo visivo pendente |
| Superfici | Intestazione, schede, cronologia/input separati | Nessuna dipendenza visiva aggiuntiva |
| Animazione | Solo cursore standard | Riproduzione al 10% non applicabile |
| Icone | Stati/ACK scritti, selezione tra parentesi quadre | Non dipende solo da colori/icone |
| Prestazioni | 500 messaggi mostrati, pagine da 100 | Memoria limitata; carico da profilare |

| Gravità | Posizione | Prima | Dopo | Motivo |
| --- | --- | --- | --- | --- |
| Media | `pkg/tui/model.go`, invio/ricarica | Possibile perdita dell'ID pendente | Ricarica attende conferma/annullamento | Evitare ritentativi ambigui |
| Media | `pkg/tui/model.go`, input | Nessuna spiegazione del limite | Mostra limite byte/controlli | Indicare il passo successivo |

Scartati effetti cromatici di connessione, animazioni frequenti e nuovo sistema di stili: testo e componenti Bubbles coprono la prova iniziale su diversi terminali.

Un PTY SSH reale ha mostrato intestazione/messaggi persistiti; i test verificano sequenze di controllo. Restano lettori di schermo, zoom, 12 colonne e revisione visiva di tutti i terminali nativi. Sono limiti documentati dell'alpha. La prova E2EE nativa successiva è in [verifica](verification.md).

// Package tui contains the same terminal interface for native and SSH access.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
)

type listed struct {
	rooms []protocol.Conversation
	err   error
}
type arrived struct {
	generation int
	event      client.Event
	open       bool
}
type sent struct {
	generation int
	message    protocol.Message
	err        error
}
type directResult struct {
	action   string
	chat     protocol.Conversation
	requests []protocol.DirectRequest
	err      error
}
type identityResult struct {
	peer        string
	fingerprint string
	verified    bool
	err         error
}
type identityBackend interface {
	Fingerprint(context.Context, string) (string, error)
	Verify(context.Context, string, string) error
	SelfFingerprint() string
}
type Model struct {
	ctx           context.Context
	backend       client.Backend
	rooms         []protocol.Conversation
	selected      int
	input         textinput.Model
	view          viewport.Model
	width, height int
	status        string
	messages      []protocol.Message
	seq           int64
	stream        <-chan client.Event
	cancel        context.CancelFunc
	generation    int
	pending       *protocol.Send
	busy          bool
	fatal         bool
	requests      []protocol.DirectRequest
	requestView   bool
	identityView  string
}

func New(ctx context.Context, b client.Backend) *Model {
	in := textinput.New()
	in.Placeholder = "Mensaje o /comando"
	in.CharLimit = 4096
	in.Focus()
	return &Model{ctx: ctx, backend: b, input: in, view: viewport.New(), width: 80, height: 24, status: "Conectando…"}
}
func (m *Model) Init() tea.Cmd { return tea.Batch(textinput.Blink, m.list()) }
func (m *Model) list() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()
		rooms, err := m.backend.List(ctx)
		return listed{rooms, err}
	}
}
func wait(ch <-chan client.Event, g int) tea.Cmd {
	return func() tea.Msg { e, ok := <-ch; return arrived{g, e, ok} }
}
func (m *Model) selectRoom() tea.Cmd {
	if m.cancel != nil {
		m.cancel()
	}
	m.generation++
	m.seq = 0
	m.messages = nil
	m.pending = nil
	m.busy = false
	m.fatal = false
	m.requestView = false
	m.identityView = ""
	m.view.SetContent("")
	if len(m.rooms) == 0 {
		m.status = "Sin conversaciones. Usa /direct USUARIO para solicitar un chat."
		return nil
	}
	ctx, cancel := context.WithCancel(m.ctx)
	m.cancel = cancel
	m.stream = m.backend.Stream(ctx, m.rooms[m.selected].ID, 0)
	m.status = "Sincronizando historial…"
	if m.rooms[m.selected].Closed {
		m.status = "Conversación cerrada · solo lectura"
	}
	return wait(m.stream, m.generation)
}
func (m *Model) directCommand() tea.Cmd {
	args := strings.Fields(m.input.Value())
	if len(args) == 0 {
		return nil
	}
	action := args[0]
	if action == "/myfingerprint" && len(args) == 1 {
		secure, ok := m.backend.(identityBackend)
		if !ok {
			m.status = "La huella requiere la TUI local cifrada."
			return nil
		}
		m.input.Reset()
		m.identityView = "Tu huella:\n" + secure.SelfFingerprint() + "\n\nCompártela con tu contacto por otro medio."
		m.renderMessages()
		m.status = "Tu huella local está lista para comparar."
		return nil
	}
	if action == "/fingerprint" || action == "/verify" {
		secure, ok := m.backend.(identityBackend)
		if !ok {
			m.status = "La verificación de identidad requiere la TUI local cifrada."
			return nil
		}
		if (action == "/fingerprint" && len(args) != 2) || (action == "/verify" && len(args) != 3) || !security.Identifier(args[1]) {
			m.status = "Usa /fingerprint USUARIO o /verify USUARIO HUELLA"
			return nil
		}
		m.busy = true
		m.input.Reset()
		m.status = "Consultando identidad…"
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
			defer cancel()
			result := identityResult{peer: args[1], verified: action == "/verify"}
			if result.verified {
				result.err = secure.Verify(ctx, args[1], args[2])
			} else {
				result.fingerprint, result.err = secure.Fingerprint(ctx, args[1])
			}
			return result
		}
	}
	if action == "/requests" && len(args) == 1 {
		m.busy = true
		m.input.Reset()
		m.status = "Cargando solicitudes…"
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
			defer cancel()
			requests, err := m.backend.DirectRequests(ctx)
			return directResult{action: action, requests: requests, err: err}
		}
	}
	if len(args) != 2 || !security.Identifier(args[1]) {
		m.status = "Usa /requests o /direct, /decline, /block, /unblock USUARIO"
		return nil
	}
	peer := args[1]
	if action != "/direct" && action != "/decline" && action != "/block" && action != "/unblock" {
		m.status = "Comando desconocido. Usa /requests o /direct, /decline, /block, /unblock USUARIO"
		return nil
	}
	m.busy = true
	m.input.Reset()
	m.status = "Procesando…"
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()
		var result directResult
		result.action = action
		switch action {
		case "/direct":
			result.chat, result.err = m.backend.Direct(ctx, peer)
		case "/decline":
			result.err = m.backend.DeclineDirect(ctx, peer)
		case "/block":
			result.err = m.backend.BlockDirect(ctx, peer)
		case "/unblock":
			result.err = m.backend.UnblockDirect(ctx, peer)
		}
		return result
	}
}
func (m *Model) send() tea.Cmd {
	if m.pending == nil {
		m.pending = &protocol.Send{ClientID: client.MessageID(), Body: m.input.Value()}
	}
	req := *m.pending
	conv := m.rooms[m.selected].ID
	g := m.generation
	m.busy = true
	m.status = "Enviando…"
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()
		msg, err := m.backend.Send(ctx, conv, req)
		return sent{g, msg, err}
	}
}
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(12, v.Width)
		m.height = max(8, v.Height)
		m.input.SetWidth(max(1, m.width-4))
		m.view.SetWidth(m.width)
		m.view.SetHeight(max(1, m.height-8))
		m.renderMessages()
		return m, nil
	case listed:
		if v.err != nil {
			m.status = "No se pudieron cargar las conversaciones. Ctrl+L reintenta."
			return m, nil
		}
		m.rooms = v.rooms
		m.selected = 0
		return m, m.selectRoom()
	case arrived:
		if v.generation != m.generation {
			return m, nil
		}
		if !v.open {
			m.status = "Conexión cerrada. Ctrl+L reconecta."
			return m, nil
		}
		if v.event.Err != nil {
			m.status = security.Display(v.event.Err.Error())
			m.fatal = v.event.Fatal
			if m.fatal {
				return m, nil
			}
		} else {
			for _, message := range v.event.Page.Messages {
				if message.Seq > m.seq {
					m.messages = append(m.messages, message)
					m.seq = message.Seq
				}
			}
			if len(m.messages) > 500 {
				m.messages = m.messages[len(m.messages)-500:]
			}
			m.renderMessages()
			m.status = fmt.Sprintf("Conectado · SYNC %d", m.seq)
			if m.rooms[m.selected].Closed {
				m.status = "Conversación cerrada · solo lectura"
			}
		}
		return m, wait(m.stream, m.generation)
	case sent:
		if v.generation != m.generation {
			return m, nil
		}
		m.busy = false
		if v.err != nil {
			m.status = "Sin confirmar. Ctrl+R reintenta el mismo mensaje; Esc cancela."
			return m, nil
		}
		m.pending = nil
		m.input.Reset()
		m.status = fmt.Sprintf("Guardado · ACK %d", v.message.Seq)
		return m, nil
	case directResult:
		m.busy = false
		if v.err != nil {
			m.status = security.Display(v.err.Error())
			return m, nil
		}
		if v.action == "/requests" {
			m.requests = v.requests
			m.requestView = true
			m.renderMessages()
			m.status = "Solicitudes · /direct USUARIO acepta · /decline USUARIO rechaza"
			return m, nil
		}
		if v.action == "/direct" && v.chat.Pending {
			m.status = "Solicitud enviada; espera la aceptación del destinatario"
			return m, nil
		}
		return m, m.list()
	case identityResult:
		m.busy = false
		if v.err != nil {
			m.status = security.Display(v.err.Error())
			return m, nil
		}
		if v.verified {
			m.identityView = ""
			m.status = "Identidad verificada para " + v.peer
			return m, m.list()
		}
		m.identityView = "Huella de " + v.peer + ":\n" + v.fingerprint + "\n\nCompárala con esa persona por otro medio. Luego escribe /verify " + v.peer + " HUELLA"
		m.renderMessages()
		m.status = "Compara la huella fuera de Karatty antes de verificar."
		return m, nil
	case tea.KeyPressMsg:
		switch v.String() {
		case "ctrl+c", "ctrl+d":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case "ctrl+l":
			if m.pending != nil || m.busy {
				m.status = "Hay un envío sin confirmar. Ctrl+R reintenta; Esc permite editar."
				return m, nil
			}
			return m, m.list()
		case "tab":
			if len(m.rooms) > 0 && !m.busy && m.pending == nil {
				m.selected = (m.selected + 1) % len(m.rooms)
				return m, m.selectRoom()
			}
			return m, nil
		case "pgup":
			m.view.PageUp()
			return m, nil
		case "pgdown":
			m.view.PageDown()
			return m, nil
		case "esc":
			if !m.busy {
				m.pending = nil
				m.status = "Edición del mensaje"
			}
			return m, nil
		case "ctrl+r":
			if !m.busy && m.pending != nil && len(m.rooms) > 0 && !m.fatal {
				return m, m.send()
			}
			return m, nil
		case "enter":
			if strings.HasPrefix(m.input.Value(), "/") && !m.busy && m.pending == nil {
				return m, m.directCommand()
			}
			if len(m.rooms) > 0 && m.rooms[m.selected].Closed {
				m.status = "Conversación cerrada · solo lectura"
				return m, nil
			}
			if m.input.Value() != "" && !security.Body(m.input.Value()) {
				m.status = "Máximo 4096 bytes; sin saltos ni caracteres de control."
				return m, nil
			}
			if !m.busy && m.pending == nil && len(m.rooms) > 0 && !m.fatal && security.Body(m.input.Value()) {
				return m, m.send()
			}
			return m, nil
		}
	}
	if !m.busy && m.pending == nil {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}
func (m *Model) renderMessages() {
	bottom := m.view.AtBottom()
	if m.identityView != "" {
		m.view.SetContent(m.identityView)
		return
	}
	if m.requestView {
		lines := []string{"Solicitudes de chat directo:"}
		if len(m.requests) == 0 {
			lines = append(lines, "No tienes solicitudes pendientes.")
		}
		for _, request := range m.requests {
			lines = append(lines, "• "+security.Display(request.Username))
		}
		lines = append(lines, "", "/direct USUARIO acepta · /decline USUARIO rechaza · /block USUARIO bloquea")
		m.view.SetContent(strings.Join(lines, "\n"))
		return
	}
	lines := make([]string, 0, len(m.messages))
	for _, msg := range m.messages {
		line := fmt.Sprintf("%6d  %s  %s", msg.Seq, security.Display(msg.Username), security.Display(msg.Body))
		lines = append(lines, ansi.Wrap(line, max(1, m.width), " "))
	}
	m.view.SetContent(strings.Join(lines, "\n"))
	if bottom {
		m.view.GotoBottom()
	}
}
func (m *Model) View() tea.View {
	var tabs []string
	for i, c := range m.rooms {
		name := security.Display(c.Name)
		if i == m.selected {
			name = "[" + name + "]"
		}
		tabs = append(tabs, name)
	}
	title := "Karatty · alpha"
	status := ansi.Truncate(m.status, m.width, "…")
	header := ansi.Truncate(strings.Join(tabs, "  "), m.width, "…")
	content := ansi.Truncate(title, m.width, "…") + "\n" + header + "\n" + strings.Repeat("─", m.width) + "\n" + m.view.View() + "\n" + strings.Repeat("─", m.width) + "\n" + m.input.View() + "\n" + status + "\n" + ansi.Truncate("/myfingerprint · /fingerprint USUARIO · /verify USUARIO HUELLA · /direct USUARIO · Ctrl+L", m.width, "…")
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

# TERMLINK (nombre provisional)

Mensajería propia desde la terminal, con un cliente nativo y acceso SSH. Repositorio privado de trabajo de karatekah. Sin integración ni código de WhatsApp o wacon-mcp.

## Sprint 1

Objetivo: dos usuarios invitados conversan entre SSH y el cliente instalable, salen y recuperan el historial desde PostgreSQL; un tercero sin membresía no puede leer, enviar ni suscribirse.

Go, monolito modular, PostgreSQL y Bubble Tea v2/Bubbles v2/Wish v2. Ambos accesos ejecutan las mismas operaciones de dominio y comparten TUI. HTTPS/WSS y SSH protegen el transporte; **el servidor puede leer los mensajes; no hay E2EE**.

La implementación y los comandos se completan durante el sprint. Consulta [las decisiones](docs/adr/0001-foundation.md), [el alcance](docs/roadmap.md) y [seguridad](docs/security.md). No hay servidor público ni servicios contratados.

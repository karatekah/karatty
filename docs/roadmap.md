# Alcance y decisiones de producto

Sprint 1: flujo funcional verificable, entorno local y CI. La alpha v0.1 requiere además recuperación, borrado de cuentas, política de datos, restauración de respaldos y revisión de límites operativos. No se abrirá un servicio en este sprint.

Decisión del usuario (2026-10-31): retención de mensajes de **30 días** para la futura alpha, con borrado de contenido e identificadores y deduplicación garantizada solo durante ese plazo. El historial ya oculta mensajes vencidos y el servidor dispone de purga por lotes, manual o programada con `--retention-30d`. La programación permanece desactivada por defecto en desarrollo. Antes de activar una alpha faltan comunicación al usuario, vencimiento/restauración de respaldos y pruebas operativas de la programación.

Horizonte original: unas 12 semanas para validar el producto, estimación sin fecha prometida. Público: desarrolladores y equipos técnicos pequeños. v0.2 perfiles, búsqueda y notificaciones; v0.3 archivos e integraciones; v0.4 bots/API; v1.0 organizaciones y facturación. Community/Pro/Teams/self-hosted son hipótesis. Precio, identidad visual y dominio siguen abiertos.

TERMLINK es un nombre provisional. Existen [getterm.link](https://www.getterm.link/) y [DimitriGeelen/termlink](https://github.com/DimitriGeelen/termlink). Hace falta investigar nombre, marca y dominio antes de publicación comercial. No se afirma disponibilidad jurídica ni se registra dominio.

Antes de prometer E2EE, evaluar un diseño auditado y Matrix. La validación comercial medirá retención, retorno e invitaciones mediante agregados, sin inspeccionar mensajes. La adecuación legal de una operación comercial queda pendiente de revisión actual de fuentes oficiales y asesoría aplicable; este repositorio no afirma cumplimiento.

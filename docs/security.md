# Seguridad y privacidad

Pruebas obligatorias: aislamiento de miembros, revocación, invitaciones, firma y reuso de desafíos, secuencia, deduplicación y persistencia. No registrar cuerpos, tokens, invitaciones, firmas ni claves privadas. Solo logs de inicio/error genéricos y contadores operativos.

Límites del sprint: mensajes hasta 4096 bytes UTF-8, sin caracteres de control ni formato invisible, páginas de 100, conexiones simultáneas acotadas, tiempo de handshake, tasas de peticiones y envío, timeout de escritura. Un consumidor lento se desconecta y puede sincronizar por secuencia. El cliente limita lo retenido en pantalla.

TLS obligatorio excepto modo de desarrollo explícito restringido a loopback. SSH utiliza clave de host persistente y el cliente debe verificarla; no desactivar StrictHostKeyChecking. Los servidores escuchan en loopback por defecto.

Pendientes antes de usuarios reales: recuperación de cuenta, clave de recuperación/segunda clave con prueba de posesión, borrado de cuentas y claves, purga a 30 días, respaldos cifrados y restauración ensayada, políticas de privacidad y retención de respaldos, gestión de secretos y revisión externa. Revocación administrativa queda disponible en el sprint. No hay E2EE.

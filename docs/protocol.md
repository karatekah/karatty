# Protocolo v1 del sprint

JSON sobre HTTPS. Todos los endpoints salvo desafío/prueba/health requieren `Authorization: Bearer SESSION_TOKEN`; no se admiten tokens en URLs. Respuestas con `Cache-Control: no-store`. Errores: 400 entrada inválida, 401 autenticación, 403 membresía, 409 ID reutilizado con contenido distinto, 429 límite. Errores internos no incluyen consultas ni datos.

| Operación | Ruta |
| --- | --- |
| Desafío de registro/login | POST /v1/auth/challenge |
| Firma SSH serializada (JSON codifica []byte como base64) | POST /v1/auth/proof |
| Cierre/revocación de sesión | DELETE /v1/session |
| Conversaciones accesibles (máximo 100 en sprint) | GET /v1/conversations |
| Chat directo con usuario existente | POST /v1/direct |
| Historial ascendente por secuencia | GET /v1/conversations/{id}/messages?after=0&limit=100 |
| Envío con client_id estable y body | POST /v1/conversations/{id}/messages |
| Eventos por secuencia, cabecera Bearer | WSS /v1/conversations/{id}/events?after=0 |

`protocol.Message` es también el ACK de persistencia: ID del servidor, client_id, conversación, remitente, username, seq, body y created_at. No significa leído ni entregado a todas las sesiones. Se devuelve después del commit. Dos remitentes pueden usar el mismo client_id; un remitente reutiliza exactamente el mismo ID al reintentar el mismo contenido en la misma conversación. La deduplicación está garantizada durante los 30 días de retención del mensaje. Al purgar un mensaje vencido se borran su contenido e identificadores; un envío posterior con el mismo client_id se considera nuevo. La purga física se ejecuta manualmente o al activar `server --retention-30d`.

`Page` contiene messages y next. Durante la sesión, el cliente guarda en memoria la última secuencia entregada y reconecta desde ella; descarta eventos ya vistos. La TUI sincroniza desde cero al abrirse de nuevo. El servidor asigna secuencias bajo bloqueo de conversación, en la transacción del INSERT. La purga puede dejar huecos en la numeración; next_seq no retrocede. El historial no entrega mensajes con más de 720 horas, incluso antes de la purga física. Historial y suscripción se autorizan incluso si el lote está vacío. No hay cola ilimitada en memoria: cada consulta devuelve 100 mensajes; consumidor lento agota el timeout de escritura y reconecta al registro durable.

SSH usa handshake de clave pública y una sesión de dominio por canal SSH. Shell con PTY abre la TUI. Comandos permitidos: `list`, `history CONVERSATION AFTER`, `send CONVERSATION CLIENT_ID` con cuerpo por stdin. No se ejecutan comandos del sistema. El registro requiere primero asociar la clave mediante CLI/invitación; un usuario ya registrado accede después desde cualquier equipo que tenga su clave y un cliente SSH.

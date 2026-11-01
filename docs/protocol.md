# Protocolo v1 del sprint

JSON sobre HTTPS. Todos los endpoints salvo desafío/prueba/health requieren `Authorization: Bearer SESSION_TOKEN`; no se admiten tokens en URLs. Respuestas con `Cache-Control: no-store`. Errores: 400 entrada inválida, 401 autenticación, 403 membresía, 409 ID reutilizado con contenido distinto, 429 límite. Errores internos no incluyen consultas ni datos.

| Operación | Ruta |
| --- | --- |
| Desafío de registro/login | POST /v1/auth/challenge |
| Firma SSH serializada (JSON codifica []byte como base64) | POST /v1/auth/proof |
| Cierre/revocación de sesión | DELETE /v1/session |
| Eliminación de la cuenta autenticada | DELETE /v1/account |
| Conversaciones accesibles (máximo 100 en sprint) | GET /v1/conversations |
| Solicitar chat directo o aceptar solicitud entrante | POST /v1/direct |
| Solicitudes directas entrantes | GET /v1/direct/requests |
| Rechazar una solicitud entrante | DELETE /v1/direct/requests/{username} |
| Bloquear usuario y cerrar chat directo | PUT /v1/direct/blocks/{username} |
| Desbloquear usuario | DELETE /v1/direct/blocks/{username} |
| Historial ascendente por secuencia | GET /v1/conversations/{id}/messages?after=0&limit=100 |
| Envío con client_id estable y body | POST /v1/conversations/{id}/messages |
| Eventos por secuencia, cabecera Bearer | WSS /v1/conversations/{id}/events?after=0 |

`protocol.Message` es también el ACK de persistencia: ID del servidor, client_id, conversación, remitente, username, seq, body y created_at. No significa leído ni entregado a todas las sesiones. Se devuelve después del commit. Dos remitentes pueden usar el mismo client_id; un remitente reutiliza exactamente el mismo ID al reintentar el mismo contenido en la misma conversación. La deduplicación está garantizada durante los 30 días de retención del mensaje. Al purgar un mensaje vencido se borran su contenido e identificadores; un envío posterior con el mismo client_id se considera nuevo. La purga física se ejecuta manualmente o al activar `server --retention-30d`.

`Page` contiene messages y next. Durante la sesión, el cliente guarda en memoria la última secuencia entregada y reconecta desde ella; descarta eventos ya vistos. La TUI sincroniza desde cero al abrirse de nuevo. El servidor asigna secuencias bajo bloqueo de conversación, en la transacción del INSERT. La purga y la eliminación de cuentas pueden dejar huecos en la numeración; next_seq no retrocede. El historial no entrega mensajes con más de 720 horas, incluso antes de la purga física. Historial y suscripción se autorizan incluso si el lote está vacío. No hay cola ilimitada en memoria: cada consulta devuelve 100 mensajes; consumidor lento agota el timeout de escritura y reconecta al registro durable.

`Conversation.closed` indica una conversación directa archivada tras la eliminación de uno de sus participantes. El otro puede leer sus propios mensajes hasta su vencimiento, pero no enviar nuevos. La eliminación borra la cuenta, sus claves, sesiones, membresías y mensajes; no borra los mensajes escritos por otras personas. Los respaldos requieren una política de vencimiento separada.

`POST /v1/direct` devuelve `pending:true` y un ID vacío cuando crea o repite una solicitud. Solo la llamada recíproca del destinatario abre el chat y devuelve su ID. Las solicitudes vencen a los 7 días; el listado no muestra solicitudes vencidas. Rechazar elimina la solicitud entrante. Bloquear elimina las solicitudes de ambas direcciones, impide nuevas solicitudes y cierra el chat directo existente; ambas partes pueden leer el historial ya autorizado hasta su vencimiento. Desbloquear permite iniciar un nuevo intercambio, sin reabrir el chat archivado.

SSH usa handshake de clave pública y una sesión de dominio por canal SSH. Shell con PTY abre la TUI. Comandos permitidos: `list`, `requests`, `direct USERNAME`, `decline USERNAME`, `block USERNAME`, `unblock USERNAME`, `history CONVERSATION AFTER`, `send CONVERSATION CLIENT_ID` con cuerpo por stdin. No se ejecutan comandos del sistema. El registro requiere primero asociar la clave mediante CLI/invitación; un usuario ya registrado accede después desde cualquier equipo que tenga su clave y un cliente SSH.

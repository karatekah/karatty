# Protocolo v1 del sprint

JSON sobre HTTPS. Todos los endpoints salvo desafío/prueba/health requieren `Authorization: Bearer SESSION_TOKEN`; no se admiten tokens en URLs. Respuestas con `Cache-Control: no-store`. Errores: 400 entrada inválida, 401 autenticación, 403 membresía, 409 ID reutilizado con contenido distinto, 429 límite. Errores internos no incluyen consultas ni datos.

La respuesta de autenticación incluye `public_mode`. El cliente local exige el flujo cifrado cuando vale `true`; el CLI legado `--legacy` se limita al servidor de invitaciones para desarrollo local.

| Operación | Ruta |
| --- | --- |
| Desafío de registro/login | POST /v1/auth/challenge |
| Firma SSH serializada (JSON codifica []byte como base64) | POST /v1/auth/proof |
| Inicio de recuperación con seudónimo y clave pública nueva | POST /v1/recovery/start |
| Firma de la nueva clave y código de recuperación | POST /v1/recovery/complete |
| Cierre/revocación de sesión | DELETE /v1/session |
| Eliminación de la cuenta autenticada | DELETE /v1/account |
| Conversaciones accesibles (máximo 100 en sprint) | GET /v1/conversations |
| Solicitar chat directo o aceptar solicitud entrante | POST /v1/direct |
| Solicitudes directas entrantes | GET /v1/direct/requests |
| Rechazar una solicitud entrante | DELETE /v1/direct/requests/{username} |
| Bloquear usuario y cerrar chat directo | PUT /v1/direct/blocks/{username} |
| Desbloquear usuario | DELETE /v1/direct/blocks/{username} |
| Subir identidad Olm y claves públicas de un solo uso firmadas (experimental) | POST /v1/e2ee/device |
| Leer identidad pública del contacto tras aceptar el directo (experimental) | GET /v1/e2ee/peer/{username} |
| Consumir una clave pública de un solo uso del contacto (experimental) | POST /v1/e2ee/peer/{username}/claim |
| Historial ascendente por secuencia | GET /v1/conversations/{id}/messages?after=0&limit=100 |
| Envío con client_id estable y body | POST /v1/conversations/{id}/messages |
| Eventos por secuencia, cabecera Bearer | WSS /v1/conversations/{id}/events?after=0 |

`protocol.Message` es también el ACK de persistencia: ID del servidor, client_id, conversación, remitente, username, seq, body y created_at. No significa leído ni entregado a todas las sesiones. Se devuelve después del commit. Dos remitentes pueden usar el mismo client_id; un remitente reutiliza exactamente el mismo ID al reintentar el mismo contenido en la misma conversación. La deduplicación está garantizada durante los 30 días de retención del mensaje. Al purgar un mensaje vencido se borran su contenido e identificadores; un envío posterior con el mismo client_id se considera nuevo. La purga física se ejecuta manualmente o al activar `server --retention-30d`.

Los endpoints `e2ee` distribuyen claves públicas según el [ADR 0003](adr/0003-e2ee-1to1.md). `POST /device` exige una firma de la clave SSH de sesión sobre `KARATTY-E2EE-IDENTITY-v1\nUSER_ID\nED25519\nCURVE25519`; cada clave de un solo uso exige firma de la identidad Olm Ed25519 sobre `KARATTY-OTK-v1\nCURVE25519\nONE_TIME_KEY`. Las claves se codifican como base64 estándar sin relleno. La identidad no puede sustituirse silenciosamente; la recuperación de cuenta la elimina. Solo un miembro de un directo aceptado puede consultar la identidad del otro y consumir una clave; una clave consumida no puede volver a activarse aunque se suba de nuevo. El cliente oficial compara la huella fuera de Karatty, cifra con Olm y guarda el ratchet en el archivo local cifrado.

`Page` contiene messages y next. Durante la sesión, el cliente guarda en memoria la última secuencia entregada y reconecta desde ella; descarta eventos ya vistos. La TUI sincroniza desde cero al abrirse de nuevo. El servidor asigna secuencias bajo bloqueo de conversación, en la transacción del INSERT. La purga y la eliminación de cuentas pueden dejar huecos en la numeración; next_seq no retrocede. El historial no entrega mensajes con más de 720 horas, incluso antes de la purga física. Historial y suscripción se autorizan incluso si el lote está vacío. No hay cola ilimitada en memoria: cada consulta devuelve 100 mensajes; consumidor lento agota el timeout de escritura y reconecta al registro durable.

`Conversation.closed` indica una conversación directa archivada tras la eliminación de uno de sus participantes. El otro puede leer sus propios mensajes hasta su vencimiento, pero no enviar nuevos. La eliminación borra la cuenta, sus claves, sesiones, membresías y mensajes; no borra los mensajes escritos por otras personas. Los respaldos requieren una política de vencimiento separada.

`POST /v1/direct` devuelve `pending:true` y un ID vacío cuando crea o repite una solicitud. Solo la llamada recíproca del destinatario abre el chat y devuelve su ID. Las solicitudes vencen a los 7 días; el listado no muestra solicitudes vencidas. Rechazar elimina la solicitud entrante. Bloquear elimina las solicitudes de ambas direcciones, impide nuevas solicitudes y cierra el chat directo existente; ambas partes pueden leer el historial ya autorizado hasta su vencimiento. Desbloquear permite iniciar un nuevo intercambio, sin reabrir el chat archivado.

El modo `server --public-signup` desactiva el registro por invitación y permite crear una cuenta con seudónimo y clave pública Ed25519, sin correo. En este modo, la API solo lista directos y rechaza historial, eventos y envíos de canales, incluso si existían antes. Los envíos a directos requieren en `body` un sobre JSON versionado con `version:1`, `session_id`, `sender_identity` y `message:{type,body}`; las claves y el cuerpo Olm se codifican en base64 estándar sin relleno. La API comprueba únicamente la **estructura**, no puede demostrar que el cuerpo sea cifrado válido. El cliente oficial produce y consume estos sobres localmente; el ensayo usa loopback y el despliegue exige la opción explícita --public-network con TLS. `auth/challenge` devuelve un desafío de firma con dominio `KARATTY-PUBLIC-REGISTER-v1`; `auth/proof` lo consume en un solo intento y exige la firma de la nueva clave. La respuesta incluye una sesión temporal y un código de recuperación de 192 bits que se muestra una sola vez. En modo invitación, el registro también entrega ese código. Para reemplazar una clave perdida, `recovery/start` recibe el seudónimo y la clave pública nueva; `recovery/complete` exige la firma de esa clave y el código guardado. Al completar, revoca todas las claves y sesiones anteriores y rota el código. Los desafíos vencen en dos minutos. La respuesta de inicio de recuperación no revela si existe el seudónimo.

El prototipo SSH usa handshake de clave pública y una sesión de dominio por canal SSH. Shell con PTY abre una TUI en el servidor. Comandos permitidos: `list`, `requests`, `direct USERNAME`, `decline USERNAME`, `block USERNAME`, `unblock USERNAME`, `history CONVERSATION AFTER`, `send CONVERSATION CLIENT_ID` con cuerpo por stdin. No se ejecutan comandos del sistema. **Este acceso no es E2EE y queda fuera de la alpha pública.** El servidor rechaza la combinación `--public-signup` y `--ssh-host-key`.

`--public-signup` exige loopback salvo que se indique también `--public-network`; esta opción requiere TLS y rechaza `--dev` y TUI SSH remota. La configuración de instalación está en [deploy.md](deploy.md). La infraestructura y la operación pública siguen pendientes.

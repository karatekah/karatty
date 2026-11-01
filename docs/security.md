# Seguridad y privacidad

Pruebas obligatorias: aislamiento de miembros, revocación, invitaciones, firma y reuso de desafíos, secuencia, deduplicación y persistencia. No registrar cuerpos, tokens, invitaciones, firmas ni claves privadas. Solo logs de inicio/error genéricos y contadores operativos.

Límites del sprint: mensajes hasta 4096 bytes UTF-8, sin caracteres de control ni formato invisible, páginas de 100, conexiones simultáneas acotadas, tiempo de handshake, tasas de peticiones y envío, timeout de escritura. Un consumidor lento se desconecta y puede sincronizar por secuencia. El cliente limita lo retenido en pantalla.

TLS obligatorio excepto modo de desarrollo explícito restringido a loopback. SSH utiliza clave de host persistente y el cliente debe verificarla; no desactivar StrictHostKeyChecking. Los servidores escuchan en loopback por defecto.

Pendientes antes de usuarios reales: configuración del correo/dominio, activación y prueba operativa de la purga a 30 días, respaldos cifrados y restauración ensayada, políticas de privacidad y retención de respaldos, gestión de secretos y revisión externa. La eliminación de cuenta y claves ya está implementada en la base activa; los respaldos quedan pendientes. El registro verificado y la recuperación con correo más código de recuperación están implementados y probados con un emisor simulado, pero aún no se ha probado la entrega real. Revocación administrativa queda disponible en el sprint. No hay E2EE.

Los límites de tasa son por proceso: API 10 peticiones/s por IP con ráfaga de 30, envío 2 mensajes/s por usuario con ráfaga de 10, SSH 1 intento/s por IP con ráfaga de 10. Mapa de límites acotado a 4096 identidades, 128 conexiones HTTP, 64 SSH y 64 suscripciones/canales SSH por proceso; tablas de desafíos con máximo 4096 y TTL de dos minutos. Hay latidos WSS cada 15 s, timeout de escritura/latido de 5 s, handshake SSH de 10 s y sesión máxima de 8 h. Revisar límites compartidos y capacidad antes de aumentar procesos o exponer a Internet.

Reabrir un chat directo existente exige conservar ambas membresías; no restaura un acceso revocado. La recuperación de acceso requiere intervención administrativa explícita. Los IDs de mensajes pendientes del cliente viven en memoria: la recuperación de una cola de salida tras cerrar el proceso queda pendiente; los scripts pueden conservar y reutilizar `--id`.

Un chat directo solo se abre con la llamada recíproca del destinatario. Las solicitudes caducan en 7 días, se limitan a 10 salientes activas por usuario y a una tasa de 1/minuto con ráfaga de 3 por proceso. El destinatario puede rechazar o bloquear. El bloqueo cierra el chat existente e impide mensajes y solicitudes nuevos; no borra los mensajes que cada participante ya podía leer.

Los códigos de correo tienen 96 bits aleatorios, vencen en 10 minutos y cada desafío admite un intento. El código de recuperación tiene 192 bits aleatorios, se almacena solo como hash y se rota al usarlo. La recuperación revoca todas las claves SSH y sesiones previas. Se admiten hasta 80 correos de registro/recuperación por día UTC; el límite y los bloqueos de diez minutos por correo se guardan en PostgreSQL. El proveedor de correo recibe la dirección y el código para entregarlo; esto debe figurar en la política de privacidad.

# ADR 0002 — Invitación y prueba de posesión

Estado: aceptado. Fecha: 2026-10-31.

El registro y login del cliente firman un desafío aleatorio de un solo uso, con dominio TERMLINK, caducidad corta y clave pública ligada en PostgreSQL. La firma SSH se verifica antes de consumir la invitación y registrar la clave. La invitación tiene hash, fecha límite y usos restantes; se consume bajo bloqueo en la misma transacción. No se envían claves privadas al servidor.

SSH verifica la posesión durante el handshake y solo admite claves previamente asociadas. Registro por CLI local con invitación, después acceso SSH sin instalar TERMLINK en las demás máquinas. Invitaciones pueden asignar una membresía inicial a un canal. El administrador opera mediante acceso local a la base, no hay endpoint público de administración.

Tokens aleatorios de sesión solo en memoria del cliente y hashes en servidor. No se guardan archivos de tokens. Claves cifradas se desbloquean con entrada oculta; el usuario protege el archivo mediante permisos del sistema. Integración con almacenes del SO, múltiples dispositivos, recuperación autoservicio y rotación asistida son backlog previo a una alpha pública.

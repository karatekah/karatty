# ADR 0002 — Invitación y prueba de posesión

Estado: aceptado. Fecha: 2026-10-31.

El registro y login del cliente firman un desafío aleatorio de un solo uso, con dominio TERMLINK, caducidad corta y clave pública ligada en PostgreSQL. La firma SSH se verifica antes de consumir la invitación y registrar la clave. La invitación tiene hash, fecha límite y usos restantes; se consume bajo bloqueo en la misma transacción. No se envían claves privadas al servidor.

SSH verifica la posesión durante el handshake y solo admite claves previamente asociadas. El prototipo permite registro por invitación y, después, acceso SSH sin instalar el cliente. Las invitaciones pueden asignar una membresía inicial a un canal. Este flujo SSH queda fuera de la primera alpha pública. El administrador opera mediante acceso local a la base, no hay endpoint público de administración.

Tokens aleatorios de sesión solo en memoria del cliente y hashes en servidor. No se guardan archivos de tokens. Claves cifradas se desbloquean con entrada oculta; el usuario protege el archivo mediante permisos del sistema. Una evolución posterior añadió registro seudónimo sin invitación y recuperación autoservicio mediante código de un solo uso; consulta [arquitectura](../architecture.md) y [protocolo](../protocol.md). Integración con almacenes del SO, múltiples dispositivos y rotación asistida siguen pendientes.

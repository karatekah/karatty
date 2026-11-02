# ADR 0001 — Núcleo compartido y PostgreSQL

[English](../../adr/0001-foundation.md) · **Español** · [Italiano](../../it/adr/0001-foundation.md)

Estado: aceptado. Fecha: 2026-10-31. Esta decisión registra la base original; las tablas de identidad y claves E2EE se incorporaron después según [ADR 0003](0003-e2ee-1to1.md).

Un módulo Go con tres ejecutables: API, gateway SSH y cliente. El gateway autentica la posesión de la clave mediante el protocolo SSH y crea una sesión del mismo núcleo. API y SSH aplican autorización por operación. Los componentes TUI se comparten mediante una interfaz pequeña.

PostgreSQL conserva usuarios, claves públicas, sesiones, invitaciones, desafíos de autenticación, conversaciones, membresías y mensajes. En el sprint inicial no se crearon tablas devices/receipts sin uso: ACK de persistencia se expresa mediante la respuesta del mensaje y su ID; confirmaciones de lectura y dispositivos quedan pendientes.

El servidor bloquea la conversación para asignar una secuencia transaccional. La unicidad por conversación/remitente/ID del cliente deduplica reintentos mientras se conserva el mensaje, durante 30 días. El mismo ID con otro contenido produce conflicto dentro de ese plazo. Al vencer, la purga borra también el ID; un envío posterior con ese ID es nuevo. No se utiliza el reloj del cliente.

Tiempo real inicial: lectura incremental del registro durable cada 250 ms mediante WSS o adaptador de núcleo. Cada lote reevalúa sesión, clave, usuario y membresía. Esto evita perder eventos entre historial y suscripción, admite varios procesos y no exige Redis. Coste: consultas periódicas y hasta 250 ms de latencia; medir antes de añadir LISTEN/NOTIFY.

Las lecturas y operaciones conservan bloqueos compartidos de autorización durante la transacción; la revocación espera las operaciones en curso y los siguientes lotes fallan. Sesiones/clave revocadas cierran la suscripción al siguiente sondeo. El gateway reevalúa también las sesiones SSH inactivas.

Paquetes users/messages/realtime se añadirán cuando tengan una responsabilidad independiente. En el sprint sus operaciones pertenecen a auth/chat/transport; no se crean directorios vacíos por convención.

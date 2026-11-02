# Cambios

## v0.1.0-alpha.1 — candidata sin publicar

- TUI local para directos 1:1 con aceptación, rechazo y bloqueo.
- Registro seudónimo mediante clave SSH Ed25519, sin correo; recuperación con código fuera del equipo.
- Cifrado Olm con vodozemac 0.11.1, huellas comparadas por otro medio y nueva verificación al cambiar la identidad.
- Archivo local con Argon2id y XChaCha20-Poly1305; cola de salida persistida antes del envío y migración tras recuperar la cuenta.
- Sobres opacos en PostgreSQL, historial por secuencia y deduplicación durante los 30 días de retención.
- Borrado de cuenta y estado local, purga de mensajes vencidos y herramientas de respaldo cifrado y restauración.
- Paquetes nativos de cliente y puente criptográfico, hashes, commit de origen y licencias de terceros.
- Alpha orientada a uso local/LAN con servidor propio, TLS y certificados con SAN de IP; guía de conversación entre dos cuentas. Dominio, VPS, buzones y servicio alojado quedan para el futuro.
- Ensayo Windows con servidor nativo y dos TUI: registro, aceptación, huellas, mensajes E2EE y reapertura del historial aprobados tras restaurar el servidor de cuarentena con Defender activo. LAN entre dos equipos físicos y detección en otras instalaciones sin comprobar.

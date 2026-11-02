# Cambios

## v0.1.0-alpha.1 — candidata sin publicar

- TUI local para directos 1:1 con aceptación, rechazo y bloqueo.
- Registro seudónimo mediante clave SSH Ed25519, sin correo; recuperación con código fuera del equipo.
- Cifrado Olm con vodozemac 0.11.1, huellas comparadas por otro medio y nueva verificación al cambiar la identidad.
- Archivo local con Argon2id y XChaCha20-Poly1305; cola de salida persistida antes del envío y migración tras recuperar la cuenta.
- Sobres opacos en PostgreSQL, historial por secuencia y deduplicación durante los 30 días de retención.
- Borrado de cuenta y estado local, purga de mensajes vencidos y herramientas de respaldo cifrado y restauración.
- Paquetes nativos de cliente y puente criptográfico, hashes, commit de origen y licencias de terceros.
- Despliegue TLS explícito sin TUI remota ni canales en el modo público. Dominio, VPS, buzones y operación pública todavía pendientes.

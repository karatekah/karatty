# Operación futura de un servicio alojado

Dominio, VPS y proveedores quedan para el futuro por decisión del usuario. No son requisitos de la alpha local/LAN descrita en [quickstart.md](quickstart.md). Esta guía prepara el despliegue en Internet; no hay infraestructura contratada ni respaldos remotos activos. El usuario eligió `karatty.dev`, pero todavía no se ha comprado ni confirmado en el registrador. Falta un correo público para soporte y privacidad; el usuario indicó que aún no tiene esos buzones. Las cuentas no requieren correo.

Nombre de trabajo: **Karatty**. El 1 de noviembre de 2026, el buscador oficial [Busca tu marca](https://pi.indecopi.gob.pe/buscatumarca/#/inicio) mostró **0 resultados en búsqueda precisa** de marcas vigentes para «Karatty» en las clases 9, 38 y 42. La búsqueda amplia devolvió 200 resultados. Esta consulta es preliminar; falta revisar posibles similitudes relevantes y confirmar dominio y marca antes de anunciar un servicio alojado con esa marca.

## Infraestructura y presupuesto

Una instancia pequeña [DigitalOcean Basic](https://www.digitalocean.com/pricing/droplets) (1 vCPU, 2 GiB RAM, 50 GB SSD) cuesta $12/mes según la tarifa consultada; verifica el precio antes de contratar. PostgreSQL y el servidor Go pueden compartirla durante una alpha de capacidad controlada. [Cloudflare R2](https://developers.cloudflare.com/r2/pricing/) ofrece una cuota gratuita de almacenamiento Standard; verifica sus límites actuales antes de contratar. El dominio `.dev` se paga anualmente; verifica precio y disponibilidad antes de comprar. Si la base o las copias superan los límites gratuitos, revisa uso y presupuesto antes de ampliar capacidad.

Comprobación del 1 de noviembre de 2026: DigitalOcean seguía mostrando **$12/mes** para esa instancia; R2 Standard incluía **10 GB-mes**, 1 millón de operaciones A y 10 millones B al mes sin cargo; [Cloudflare Registrar](https://pricing.registrar.cloudflare.com/) mostraba **$12.20/año de renovación** para `.dev` (el precio inicial puede ser promocional). La estimación recurrente base es $12 + $12.20/12 ≈ **$13.02/mes**, antes de impuestos, excedentes y otros servicios. El [RDAP del registro de `.dev`](https://pubapi.registry.google/rdap/domain/karatty.dev) respondió 404 para `karatty.dev`; eso sugiere que no estaba registrado en la consulta, pero **no confirma que el registrador lo ofrezca ni reserva el nombre**. Comprueba la compra y el precio final antes de cerrar la marca.

Para recibir consultas en `support@karatty.dev` y `privacy@karatty.dev`, [Cloudflare Email Routing](https://developers.cloudflare.com/email-service/platform/pricing/) ofrece reenvío entrante en el plan gratuito. [No permite responder desde el dominio reenviado](https://developers.cloudflare.com/email-service/reference/postmaster/#sending-or-replying-to-an-email-from-your-cloudflare-domain); la respuesta saldría del buzón de destino. El titular debe elegir y verificar ese buzón y la dirección pública antes de publicar la política. Las cuentas de Karatty no necesitan correo.

## Respaldos

Usa un bucket R2 privado y dedicado. Configura `rclone` con credenciales limitadas a ese bucket y `BACKUP_REMOTE=r2:NOMBRE_DEL_BUCKET/db-backups`. Genera una identidad `age` en un equipo distinto al VPS, guarda la clave privada fuera del VPS y configura solo su receptor público `age1…` como `BACKUP_RECIPIENT`. El script [backup.sh](../scripts/backup.sh) ejecuta `pg_dump -Fc`, cifra el flujo antes de subirlo, retiene el archivo cifrado local solo durante la subida y elimina objetos remotos con más de 3 días. Programa una ejecución diaria con `systemd` y `Persistent=true`.

Configura también en R2 una [regla de ciclo de vida](https://developers.cloudflare.com/r2/buckets/object-lifecycles/) para el prefijo `db-backups/` que caduque los objetos a los **3 días**. Verifica que la regla aparezca aplicada y que no haya versionado u otras copias fuera de esta política. Cloudflare indica que la eliminación suele ocurrir dentro de las 24 horas posteriores al vencimiento, pero puede demorarse más. Por eso la política pública de hasta 7 días adicionales en respaldos requiere monitorear el borrado real y revisar cualquier objeto que supere 4 días; no se puede afirmar una garantía física exacta basándose solo en el ciclo de vida del proveedor.

Antes de abrir el servicio, sube un respaldo y restaura uno en una base vacía cuyo nombre comience con `termlink_restore_`. En un equipo con la clave privada `age`, configura `BACKUP_OBJECT`, `AGE_IDENTITY_FILE` y `RESTORE_DATABASE_URL`, y ejecuta [restore-check.sh](../scripts/restore-check.sh). El script rehúsa bases con otro nombre o tablas existentes. Confirma que las filas esperadas se restauraron y que el servicio puede leerlas. Repite la restauración regularmente y tras cambios de esquema. Nunca guardes la identidad privada `age` en el bucket ni en el VPS.

Ensayo local del 31 de octubre de 2026: `backup.sh` cifró y subió una copia al backend local de `rclone`; `restore-check.sh` la restauró en una base desechable y verificó 3 usuarios y 9 mensajes. Una copia marcada con cinco días de antigüedad se eliminó al ejecutar el respaldo siguiente. Este ensayo no verifica R2 ni su regla de ciclo de vida.

## Condiciones antes de abrir un servicio alojado

1. Confirmar la disponibilidad y el precio de `karatty.dev` en el registrador, comprarlo, cerrar la revisión de marca y preparar correo público de soporte y privacidad.
2. Implementar y revisar E2EE para los directos 1:1 en la TUI local. Impedir mensajes en claro, canales y TUI SSH remota en el servicio público. Verificar las claves de los contactos y ensayar pérdida/rotación de claves.
3. Crear VPS y bucket privados, configurar DNS, TLS y acceso administrativo. Mantener credenciales fuera del repositorio.
4. Completar y revisar el [borrador de privacidad](privacy-draft.md), publicar política y condiciones con identidad y contacto del titular, retención de mensajes de 30 días y ventana de respaldos; completar las gestiones de datos personales aplicables en Perú.
5. Probar registro seudónimo, recuperación con código, aceptación de directos, bloqueo, borrado de cuenta, purga horaria, respaldo, restauración y borrado remoto en el entorno contratado.
6. Partir de una [versión publicada y verificada](release.md) y hacer una prueba de instalación externa contra el servicio alojado.

No anuncies el servicio alojado hasta registrar estas comprobaciones con fecha y resultado. La distribución del software local/LAN tiene su propia lista en [release.md](release.md).

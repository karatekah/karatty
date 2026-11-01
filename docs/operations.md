# Operación propuesta para la alpha pública

Esta guía prepara el despliegue; no hay infraestructura contratada ni respaldos remotos activos. El dominio y un correo público para soporte y privacidad deben fijarse antes de ejecutar el despliegue. Las cuentas no requieren correo.

Nombre de trabajo: **Karatty**. El 1 de noviembre de 2026, el buscador oficial [Busca tu marca](https://pi.indecopi.gob.pe/buscatumarca/#/inicio) mostró **0 resultados en búsqueda precisa** de marcas vigentes para «Karatty» en las clases 9, 38 y 42. La búsqueda amplia devolvió 200 resultados. Esta consulta es preliminar; falta revisar posibles similitudes relevantes y confirmar dominio y marca antes de publicar.

## Infraestructura y presupuesto

Una instancia pequeña [DigitalOcean Basic](https://www.digitalocean.com/pricing/droplets) (1 vCPU, 2 GiB RAM, 50 GB SSD) cuesta $12/mes según la tarifa consultada; verifica el precio antes de contratar. PostgreSQL y el servidor Go pueden compartirla durante una alpha de capacidad controlada. [Cloudflare R2](https://developers.cloudflare.com/r2/pricing/) ofrece una cuota gratuita de almacenamiento Standard; verifica sus límites actuales antes de contratar. El dominio `.dev` se paga anualmente; verifica precio y disponibilidad antes de comprar. Si la base o las copias superan los límites gratuitos, revisa uso y presupuesto antes de ampliar capacidad.

## Respaldos

Usa un bucket R2 privado y dedicado. Configura `rclone` con credenciales limitadas a ese bucket y `BACKUP_REMOTE=r2:NOMBRE_DEL_BUCKET/db-backups`. Genera una identidad `age` en un equipo distinto al VPS, guarda la clave privada fuera del VPS y configura solo su receptor público `age1…` como `BACKUP_RECIPIENT`. El script [backup.sh](../scripts/backup.sh) ejecuta `pg_dump -Fc`, cifra el flujo antes de subirlo, retiene el archivo cifrado local solo durante la subida y elimina objetos remotos con más de 3 días. Programa una ejecución diaria con `systemd` y `Persistent=true`.

Configura también en R2 una [regla de ciclo de vida](https://developers.cloudflare.com/r2/buckets/object-lifecycles/) para el prefijo `db-backups/` que caduque los objetos a los **3 días**. Verifica que la regla aparezca aplicada y que no haya versionado u otras copias fuera de esta política. Cloudflare indica que la eliminación suele ocurrir dentro de las 24 horas posteriores al vencimiento, pero puede demorarse más. Por eso la política pública de hasta 7 días adicionales en respaldos requiere monitorear el borrado real y revisar cualquier objeto que supere 4 días; no se puede afirmar una garantía física exacta basándose solo en el ciclo de vida del proveedor.

Antes de abrir el servicio, sube un respaldo y restaura uno en una base vacía cuyo nombre comience con `termlink_restore_`. En un equipo con la clave privada `age`, configura `BACKUP_OBJECT`, `AGE_IDENTITY_FILE` y `RESTORE_DATABASE_URL`, y ejecuta [restore-check.sh](../scripts/restore-check.sh). El script rehúsa bases con otro nombre o tablas existentes. Confirma que las filas esperadas se restauraron y que el servicio puede leerlas. Repite la restauración regularmente y tras cambios de esquema. Nunca guardes la identidad privada `age` en el bucket ni en el VPS.

Ensayo local del 31 de octubre de 2026: `backup.sh` cifró y subió una copia al backend local de `rclone`; `restore-check.sh` la restauró en una base desechable y verificó 3 usuarios y 9 mensajes. Una copia marcada con cinco días de antigüedad se eliminó al ejecutar el respaldo siguiente. Este ensayo no verifica R2 ni su regla de ciclo de vida.

## Puertas de publicación

1. Elegir nombre y comprobar marca/dominio; contratar el `.dev` y preparar correo público de soporte y privacidad.
2. Implementar y revisar E2EE para los directos 1:1 en la TUI local. Impedir mensajes en claro, canales y TUI SSH remota en el servicio público. Verificar las claves de los contactos y ensayar pérdida/rotación de claves.
3. Crear VPS y bucket privados, configurar DNS, TLS y acceso administrativo. Mantener credenciales fuera del repositorio.
4. Completar y revisar el [borrador de privacidad](privacy-draft.md), publicar política y condiciones con identidad y contacto del titular, retención de mensajes de 30 días y ventana de respaldos; completar las gestiones de datos personales aplicables en Perú.
5. Probar registro seudónimo, recuperación con código, aceptación de directos, bloqueo, borrado de cuenta, purga horaria, respaldo, restauración y borrado remoto en el entorno contratado.
6. Seguir la [guía de publicación](release.md): revisar historial, licencia y documentación; publicar el código fuente y los binarios con hashes en la misma versión y hacer una prueba de instalación externa.

No anuncies la alpha hasta que estas comprobaciones estén registradas con fecha y resultado.

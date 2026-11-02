# Instalar el servicio alpha

Estado: procedimiento preparado; todavía no se ha ejecutado en un VPS ni existe un servicio público. Requiere comprar `karatty.dev`, acceso administrativo al VPS, un bucket privado y el contacto público del operador. No se deben anunciar registros hasta completar las pruebas de [publicación](release.md).

## Configuración prevista

Una instancia de 1 vCPU y 2 GiB con Ubuntu 24.04 LTS, PostgreSQL 17 y el servidor Go en la misma máquina. Ubuntu 24.04 tiene [mantenimiento estándar hasta mayo de 2029](https://ubuntu.com/about/release-cycle). El cliente se conectará directamente por TLS a `https://chat.karatty.dev`; no se usa un proxy que oculte la dirección del cliente a los límites de tasa. El gateway SSH de mensajería no se instala. SSH administrativo del VPS se limita a las claves del operador.

En DNS, crear el registro A de `chat.karatty.dev` hacia la IPv4 del VPS, sin proxy. No publicar AAAA hasta configurar y probar la escucha y el firewall IPv6. Abrir 443/TCP para la API y 80/TCP para los desafíos HTTP de Certbot; restringir 22/TCP al acceso administrativo. PostgreSQL debe escuchar solo en loopback y usar autenticación SCRAM, sin puerto público.

## Base y usuario de servicio

Instalar PostgreSQL 17 y sus herramientas siguiendo el [repositorio oficial de PostgreSQL para Ubuntu](https://www.postgresql.org/download/linux/ubuntu/). Comprobar `postgres --version` y `pg_dump --version`; usar la misma versión mayor para copia y restauración. Crear el rol `karatty` sin privilegios de superusuario ni creación de roles, y una base `karatty` de su propiedad. Asignar su contraseña mediante `\password karatty` en `psql`, sin escribirla en el historial del shell. El rol necesita crear las tablas y migraciones dentro de su propia base.

```sh
sudo useradd --system --home /var/lib/karatty --shell /usr/sbin/nologin karatty
sudo install -d -o root -g karatty -m 0750 /etc/karatty /etc/karatty/tls
sudo install -d -o root -g root -m 0755 /opt/karatty
sudoedit /etc/karatty/server.env
sudo chmod 600 /etc/karatty/server.env
```

El archivo de entorno contiene una sola variable, con la contraseña codificada para URL cuando tenga caracteres reservados:

```dotenv
DATABASE_URL=postgres://karatty:CONTRASENA_CODIFICADA@127.0.0.1:5432/karatty?sslmode=disable
```

`sslmode=disable` se limita a la conexión local al mismo VPS. No copiar este valor a una base remota. No incluir archivos de entorno ni certificados privados en Git.

## Binario y TLS

Desde el commit candidato con CI verde, usar Go 1.27.3 para construir `go build -trimpath -o bin/karatty-server ./cmd/server`. Instalar el binario y la unidad incluidos en este repositorio:

```sh
sudo install -m 0755 bin/karatty-server /opt/karatty/server
sudo install -m 0644 deployments/systemd/karatty.service /etc/systemd/system/karatty.service
sudo install -d -m 0755 /etc/letsencrypt/renewal-hooks/deploy
sudo install -m 0755 deployments/renew-certificate.sh /etc/letsencrypt/renewal-hooks/deploy/karatty
```

Instalar Certbot y emitir un certificado para `chat.karatty.dev` con el método standalone, después de comprobar DNS y el puerto 80. El hook copia certificado y clave con permisos `root:karatty 0640`; solo actúa sobre ese nombre. Si es la primera emisión y el hook no se ejecutó, invocarlo con `RENEWED_DOMAINS=chat.karatty.dev` y `RENEWED_LINEAGE` apuntando al directorio que Certbot acaba de crear. Ensayar la renovación de Certbot antes de continuar. El reinicio renueva el certificado en memoria del servidor y los clientes reconectan desde su secuencia.

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now karatty.service
sudo systemctl status karatty.service
curl --fail https://chat.karatty.dev/healthz
```

La unidad exige `--public-signup --public-network`, certificado y clave TLS. Ese modo rechaza canales, sobres de texto legible y TUI SSH remota; también activa la purga de mensajes al arrancar y cada hora. El servidor conserva los metadatos necesarios para entregar mensajes. La estructura del sobre no demuestra que un cliente modificado haya cifrado correctamente; las pruebas se refieren al cliente oficial.

## Copias y restauración

Instalar `age`, `rclone` y `pg_dump`. Configurar un bucket R2 privado dedicado y una credencial limitada a ese bucket. Guardar el archivo `rclone.conf` en `/etc/karatty/rclone.conf` con propietario `root:karatty` y permisos 0640. En `/etc/karatty/backup.env` (0600, leído por systemd), definir:

```dotenv
BACKUP_RECIPIENT=age1RECEPTOR_PUBLICO
BACKUP_REMOTE=r2:BUCKET_PRIVADO/db-backups
RCLONE_CONFIG=/etc/karatty/rclone.conf
```

La identidad privada `age` se guarda fuera del VPS. Instalar el script y las unidades:

```sh
sudo install -m 0755 scripts/backup.sh /opt/karatty/backup.sh
sudo install -m 0644 deployments/systemd/karatty-backup.service deployments/systemd/karatty-backup.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start karatty-backup.service
sudo systemctl enable --now karatty-backup.timer
sudo systemctl list-timers karatty-backup.timer
```

Aplicar también la regla de expiración remota a tres días y comprobar el borrado real, como explica [operación](operations.md). Restaurar una copia mediante `scripts/restore-check.sh` en una base vacía de nombre `termlink_restore_…`, en un equipo con la identidad privada. Verificar el contenido cifrado, las migraciones, el registro y la autorización después de restaurar. No probar una restauración encima de la base activa.

## Antes de anunciar

Desde otra red, instalar cada paquete descargado y probar dos cuentas nuevas: registro, códigos guardados fuera del equipo, aceptación, comparación de huellas, envío en ambas direcciones, desconexión, reinicio, bloqueo, recuperación y borrado. Comprobar que la API y la base reciben sobres cifrados, que los mensajes vencidos no aparecen y que se borran sus filas. Probar una copia, su restauración y la eliminación remota de objetos caducados. Completar identidad, contacto y política del operador; registrar fecha y resultado de cada comprobación en la guía de publicación.

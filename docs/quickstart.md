# Entorno local

Requisitos: Go 1.27.3, Rust 1.99.1, PostgreSQL 17.11 (o Docker con Compose) y cliente OpenSSH. Los ejemplos levantan PostgreSQL con Docker; si ya tienes una instancia nativa, omite `docker compose ... up -d db` y ajusta `DATABASE_URL` a su host, puerto, usuario, contraseña y base. Compila sin CGo. Las dependencias exactas están en go.mod/go.sum y crypto/Cargo.lock. No se instala ni se publica un servidor remoto.

## Windows / PowerShell

Desde la raíz del repositorio:

```powershell
$env:TERMLINK_DB_PASSWORD = 'termlink-local-development-only'
docker compose -f deployments/compose.yml up -d db
$env:DATABASE_URL = "postgres://termlink:$env:TERMLINK_DB_PASSWORD@127.0.0.1:5432/termlink?sslmode=disable"
New-Item -ItemType Directory -Force bin,.local | Out-Null
go build -o bin/server.exe ./cmd/server
go build -o bin/ssh-gateway.exe ./cmd/ssh-gateway
go build -o bin/karatty.exe ./cmd/termlink
cargo build --locked --release --manifest-path crypto/Cargo.toml
Copy-Item crypto/target/release/karatty-crypto.exe bin/karatty-crypto.exe
go run ./scripts/devcert.go
./bin/server.exe --addr 127.0.0.1:8443 --tls-cert .local/tls/server.crt --tls-key .local/tls/server.key --ssh-host-key .local/ssh/ssh_host_ed25519
```

El servidor aplica las migraciones de esquema al arrancar. En otra PowerShell con el mismo DATABASE_URL:

```powershell
$channel = ./bin/server.exe --channel general
$inviteAlice = ./bin/server.exe --invite --conversation $channel --uses 1 --ttl 24h
$inviteBob = ./bin/server.exe --invite --conversation $channel --uses 1 --ttl 24h
$inviteEve = ./bin/server.exe --invite --uses 1 --ttl 24h
ssh-keygen -t ed25519 -f .local/alice
ssh-keygen -t ed25519 -f .local/bob
ssh-keygen -t ed25519 -f .local/eve
$inviteAlice | ./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice --username alice --invitation-stdin
$inviteBob | ./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/bob --username bob --invitation-stdin
$inviteEve | ./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/eve --username eve --invitation-stdin
./bin/karatty.exe chat --legacy --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice
```

Las frases de claves cifradas se piden de forma oculta en la terminal, incluso al pasar una invitación por entrada estándar. Los permisos de Windows deben permitir acceso a las claves solo a su propietario; no se cambian ACL globales. Las invitaciones se pasan mediante variables/entrada estándar para no introducirlas literalmente en el historial de comandos. No compartirlas en logs.

En otra terminal, Bob entra por SSH:

```powershell
ssh-keygen -lf .local/ssh/ssh_host_ed25519.pub
ssh -t -p 2222 -i .local/bob chat@127.0.0.1
```

Compara la huella mostrada por SSH con la del host local antes de aceptarla. Conserva la clave de host entre reinicios. No desactives su verificación. `Ctrl+C` sale; `Tab` cambia de conversación; `PgUp/PgDn` desplaza el historial conservado en pantalla; `Ctrl+R` reintenta un envío sin confirmar con el mismo ID; `Esc` lo devuelve a edición; `Ctrl+L` vuelve a cargar conversaciones y sincroniza desde secuencia cero.

El cliente conserva como máximo 500 mensajes en pantalla; el historial disponible de los últimos 30 días se consulta por páginas:

```powershell
./bin/karatty.exe history --legacy --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice --conversation $channel --after 0
ssh -p 2222 -i .local/eve chat@127.0.0.1 history $channel 0
```

El primero devuelve hasta 100 mensajes y `next`; úsalo como `--after` para la siguiente página. Eve recibe rechazo porque su invitación no da membresía al canal. Eve tampoco puede enviar ni suscribirse a sus eventos. La suite automatizada verifica los tres casos.

Solicitar un chat directo: `karatty direct --peer bob` con los mismos flags de conexión. Devuelve `pending:true` y ningún ID. Bob ve la solicitud con `karatty requests` y la acepta con `karatty direct --peer alice`; entonces ambos ven el chat al recargar. Bob puede rechazarla con `karatty decline --peer alice` o bloquear a Alice con `karatty block --peer alice`; `karatty unblock --peer alice` permite recibir una solicitud nueva. En la TUI se usan `/requests`, `/direct USUARIO`, `/decline USUARIO`, `/block USUARIO` y `/unblock USUARIO`. Las solicitudes vencen a los 7 días. La identidad de la sesión siempre procede de la clave, no del nombre `chat` de SSH.

## Ensayo local del registro público

Detén el servidor de ejemplo y reinícialo con `--public-signup`, sin `--ssh-host-key`. Este modo permite registro con seudónimo y clave Ed25519, sin correo ni invitación; activa automáticamente la purga horaria de mensajes vencidos. El cliente TUI local cifra los directos, pero el servicio continúa restringido a loopback mientras se completan las pruebas y la operación. Después de crear una clave Ed25519, ejecuta:

```powershell
./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice --username alice --public
```

El cliente muestra un código de recuperación una sola vez: guárdalo fuera del equipo. Después pide y confirma una frase secreta local de al menos 12 caracteres. Esta frase no sale del equipo; protege el historial legible y el estado del cifrado. Si se pierde, el archivo local no se puede abrir. El código de recuperación permite reemplazar la clave SSH, pero no restaura el historial E2EE anterior. Si se pierde la clave SSH, genera una nueva y ejecuta `karatty recover --key NUEVA_CLAVE --username alice` con los mismos flags de conexión; se pedirá el código guardado fuera del equipo.

Para probar dos usuarios, registra también a Bob en otro directorio de estado (`--state-dir RUTA_ABSOLUTA`) y con otra clave SSH. Alice solicita `karatty direct --peer bob`; Bob acepta con `karatty direct --peer alice`. Cada persona abre `karatty chat` con su propia clave y servidor. En la TUI, `/myfingerprint` muestra la huella propia para compartir por otro medio; `/fingerprint USUARIO` muestra la del contacto. Compara las dos fuera de Karatty y escribe `/verify USUARIO HUELLA` en ambos equipos. Solo entonces se habilitan los envíos y la lectura de ese directo. El archivo local se desbloquea con la frase secreta en cada arranque.

## Linux y macOS

Mismos comandos Go, Docker y SSH. Usa `export` y ejecutables sin `.exe`:

```sh
export TERMLINK_DB_PASSWORD=termlink-local-development-only
export DATABASE_URL="postgres://termlink:$TERMLINK_DB_PASSWORD@127.0.0.1:5432/termlink?sslmode=disable"
docker compose -f deployments/compose.yml up -d db
mkdir -p bin .local
chmod 700 .local
go build -o bin/server ./cmd/server
go build -o bin/karatty ./cmd/termlink
cargo build --locked --release --manifest-path crypto/Cargo.toml
cp crypto/target/release/karatty-crypto bin/karatty-crypto
go run ./scripts/devcert.go
bin/server --addr 127.0.0.1:8443 --tls-cert .local/tls/server.crt --tls-key .local/tls/server.key --ssh-host-key .local/ssh/ssh_host_ed25519
# En otra terminal con DATABASE_URL:
channel=$(bin/server --channel general)
invite=$(bin/server --invite --conversation "$channel" --uses 1)
ssh-keygen -t ed25519 -f .local/alice
printf '%s' "$invite" | bin/karatty register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice --username alice --invitation-stdin
bin/karatty chat --legacy --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice
```

Las claves Unix se crean con modo 0600 dentro de directorios 0700. Si se copian de otra máquina, revisa propietario y permisos. La CA de desarrollo solo se confía en este cliente mediante `--ca-cert`; no se añade al almacén global. Caduca a los 30 días. La CA privada no se guarda. Para renovarla, mueve la carpeta TLS antigua a una copia privada y vuelve a generar certificados.

## Aplicación completa en Docker

Genera antes los certificados con `go run ./scripts/devcert.go`. En Linux/macOS exporta `TERMLINK_UID=$(id -u)` y `TERMLINK_GID=$(id -g)` para que el contenedor pueda leer los archivos privados de desarrollo. Windows usa los valores predeterminados de Compose. Después:

```sh
docker compose -f deployments/compose.yml --profile app up -d --build
docker compose -f deployments/compose.yml exec app /server --channel general
```

API TLS en 127.0.0.1:8443, SSH en 127.0.0.1:2222. El contenedor y los procesos nativos comparten el mismo esquema y reglas; elige uno para evitar colisiones de puertos. Las claves SSH del host persisten en `.local/ssh`; la base en un volumen. `docker compose -f deployments/compose.yml down` detiene el entorno conservando sus datos. No uses `down -v` si quieres conservar el historial. Para iniciar solo el gateway por separado: `go run ./cmd/ssh-gateway --host-key .local/ssh/ssh_host_ed25519`.

## Verificación y administración

```powershell
$env:TERMLINK_TEST_DATABASE_URL = $env:DATABASE_URL
./scripts/check.ps1
# Linux/macOS: export TERMLINK_TEST_DATABASE_URL="$DATABASE_URL"
# go mod verify; go vet ./...; go test -race -count=1 ./...; go build ./cmd/...
```

Las pruebas crean y eliminan un esquema aleatorio dentro de la base indicada. Utiliza una base local de desarrollo con permiso CREATE SCHEMA. No apuntes a datos de producción. Sin variable, las pruebas unitarias se ejecutan y las de integración se omiten explícitamente; CI exige PostgreSQL.

Operaciones administrativas locales (requieren DATABASE_URL): `server --revoke-key 'SHA256:…'` y `server --revoke-member USER_ID --conversation CONVERSATION_ID`. El ID de usuario aparece al registrar; la huella de su clave se consulta con `ssh-keygen -lf ARCHIVO.pub`. La revocación no elimina el historial: bloquea nuevas operaciones y termina las suscripciones al siguiente sondeo (250 ms más el tiempo de consulta). Los datos ya entregados no se pueden retirar de la pantalla remota.

El titular puede eliminar su cuenta con `karatty delete-account --server URL --ca-cert CA --key CLAVE --confirm-delete`. La operación es irreversible: borra su usuario, claves, sesiones, membresías y mensajes. Conserva los mensajes escritos por otros participantes; los chats directos afectados quedan en modo solo lectura para ellos. Los datos de respaldos seguirán su vencimiento independiente cuando se configure y compruebe antes de la alpha pública.

Retención de mensajes: el historial oculta los mensajes después de 720 horas y conserva sus huecos de secuencia. `server --purge-expired-messages` elimina una vez todos los mensajes vencidos y muestra el número borrado. Para programar la purga al arrancar y cada hora, añade `--retention-30d` al comando de inicio de `server`; está desactivada por defecto en desarrollo. La purga borra también los identificadores de deduplicación, cuya garantía dura 30 días. Los respaldos necesitan una política de vencimiento separada antes de operar una alpha.

Endpoints operativos: `GET /healthz` comprueba PostgreSQL; `GET /metrics` expone contadores sin contenidos y requiere una sesión Bearer válida además de una conexión loopback. Con el servidor nativo, ejecuta `karatty metrics` desde la misma máquina y con los mismos flags `--server`, `--ca-cert` y `--key` que `karatty history`. No confiar en cabeceras de proxy para identidad o dirección. En contenedor accedido por red puente `/metrics` puede devolver 403; publica métricas mediante un adaptador privado antes de operar en un VPS.

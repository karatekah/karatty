# Karatty en tu equipo o en una red local

[English](../quickstart.md) · **Español** · [Italiano](../it/quickstart.md)

La alpha usa un servidor dentro de tu equipo o LAN, PostgreSQL y clientes TUI locales. No requiere dominio, VPS, correo ni acceso desde Internet. Ambos participantes deben poder alcanzar el mismo servidor. Cada cuenta utiliza un dispositivo.

Usa `--public-signup`: registro seudónimo sin correo, directos 1:1 y sobres E2EE. El nombre del flag no obliga a exponer el servidor en Internet. No uses `--legacy`, invitaciones, canales ni el gateway SSH para esta alpha.

## 1. Compilar o instalar

Para compilar necesitas Go 1.27.4, Rust 1.99.1 y OpenSSH. Solo el servidor necesita PostgreSQL 17.11, nativo o en Docker. Los clientes pueden usar los dos ejecutables de un [paquete verificado](install.md), sin Go, Rust ni PostgreSQL. El paquete del cliente no incluye el servidor.

Obtén el código con `git clone https://github.com/karatekah/karatty.git` y entra con `cd karatty`. También puedes descargar el código de la etiqueta desde la Release.

PowerShell, desde la raíz del repositorio:

```powershell
New-Item -ItemType Directory -Force bin,.local | Out-Null
go build -trimpath -o bin/server.exe ./cmd/server
go build -trimpath -o bin/karatty.exe ./cmd/termlink
cargo build --locked --release --manifest-path crypto/Cargo.toml
Copy-Item crypto/target/release/karatty-crypto.exe bin/karatty-crypto.exe
```

En Linux/macOS:

```sh
mkdir -p bin .local
chmod 700 .local
go build -trimpath -o bin/server ./cmd/server
go build -trimpath -o bin/karatty ./cmd/termlink
cargo build --locked --release --manifest-path crypto/Cargo.toml
cp crypto/target/release/karatty-crypto bin/karatty-crypto
```

Los ejemplos siguientes usan PowerShell. En Linux/macOS usa `export VARIABLE=valor` y ejecutables sin `.exe`; los flags y el flujo en la TUI son iguales.

## 2. Iniciar PostgreSQL

Si ya tienes PostgreSQL, crea una base dedicada y configura `DATABASE_URL`. No apuntes pruebas a producción. El servidor aplica migraciones al arrancar:

```powershell
$env:DATABASE_URL = 'postgres://USUARIO:CONTRASENA_CODIFICADA@127.0.0.1:5432/karatty?sslmode=disable'
```

Codifica para URL los caracteres reservados de la contraseña. `sslmode=disable` solo corresponde aquí a PostgreSQL en el mismo equipo. Mantén credenciales fuera de Git y de capturas públicas.

Como alternativa de desarrollo, Docker Compose crea la base `termlink` y publica PostgreSQL solo en loopback:

```powershell
$env:TERMLINK_DB_PASSWORD = 'termlink-local-development-only'
docker compose -f deployments/compose.yml up -d db
$env:DATABASE_URL = "postgres://termlink:$env:TERMLINK_DB_PASSWORD@127.0.0.1:5432/termlink?sslmode=disable"
```

La contraseña es exclusivamente de ejemplo local. No inicies el perfil `app` de Compose: conserva el prototipo por invitaciones y SSH para sus pruebas.

## 3. Elegir local o LAN e iniciar con TLS

**Solo este equipo:**

```powershell
go run ./scripts/devcert.go
./bin/server.exe --addr 127.0.0.1:8443 --tls-cert .local/tls/server.crt --tls-key .local/tls/server.key --public-signup
```

Ambos clientes usarán `https://127.0.0.1:8443` y `.local/tls/ca.crt`. Deja abierta la terminal del servidor.

**Red local:** busca la IPv4 privada del servidor con `ipconfig` (Windows) o `ip addr` (Linux). Sustituye `192.168.1.50` en ambos comandos por esa IP:

```powershell
go run ./scripts/devcert.go --out .local/tls-lan --ip 192.168.1.50
./bin/server.exe --addr 192.168.1.50:8443 --tls-cert .local/tls-lan/server.crt --tls-key .local/tls-lan/server.key --public-signup --public-network
```

`--public-network` permite escuchar fuera de loopback, también en LAN; exige TLS y el modo E2EE, sin `--dev`. Ambos clientes deben usar `https://192.168.1.50:8443`, incluso el instalado en el servidor. Elige una URL estable desde el primer registro: el estado cifrado se vincula a esa URL; cambiarla puede crear otra identidad local. Una reserva DHCP ayuda a conservar la IP.

Entrega **solo `ca.crt`** a cada participante por un medio de confianza y compara su SHA-256 fuera de Karatty (`Get-FileHash RUTA/ca.crt -Algorithm SHA256` o `sha256sum RUTA/ca.crt`). El cliente confía en ella con `--ca-cert`, sin instalarla globalmente. `server.key` se queda en el servidor; cada cliente conserva su propia clave SSH y estado. No compartas carpetas de estado entre equipos.

El firewall debe permitir TCP 8443 desde los clientes de la subred privada. Si necesitas una regla, limítala a esa subred y al perfil privado; no desactives el firewall. Las redes de invitados pueden aislar dispositivos. No abras PostgreSQL ni reenvíes puertos en el router. Desde otro equipo Windows comprueba alcance con `Test-NetConnection 192.168.1.50 -Port 8443`.

Los certificados caducan a los 30 días; la clave privada de la CA no se guarda. El generador rechaza sobrescribir certificados existentes. Para renovar o cambiar IP, usa otro `--out`, reinicia el servidor y redistribuye y comprueba la CA. Nunca desactives la validación TLS.

## 4. Registrar dos personas

En otra terminal, Alice crea su clave dedicada y se registra:

```powershell
ssh-keygen -t ed25519 -f .local/alice
./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice --username alice --public
./bin/karatty.exe chat --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/alice
```

Bob usa otra terminal y otra clave:

```powershell
ssh-keygen -t ed25519 -f .local/bob
./bin/karatty.exe register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/bob --username bob --public
./bin/karatty.exe chat --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/bob
```

En LAN sustituye en **todos** los comandos `--server` y `--ca-cert` por la URL de LAN y la ruta a la CA recibida. Cada persona genera la clave en su equipo; no copia la de otra. Crea primero la carpeta para tu clave si no existe. Con un paquete extraído ejecuta `./karatty.exe` y conserva `karatty-crypto.exe` junto a él.

Guarda fuera del equipo el código de recuperación mostrado una vez. Después elige y confirma una frase secreta local independiente de la frase de la clave SSH. El archivo se desbloquea cada vez que abres el chat. Las claves distintas separan el estado de Alice y Bob incluso en una misma máquina. `--state-dir RUTA_ABSOLUTA` permite elegir otro directorio; repítelo en cada comando si lo usas.

## 5. Aceptar, verificar y conversar

Dentro de ambas TUI:

```text
Alice: /direct bob
Bob:   /requests
Bob:   /direct alice
Ambos: Ctrl+L para recargar; Tab para seleccionar el directo.
Ambos: /myfingerprint
       Comparten y comparan huellas por otro medio de confianza.
Alice: /fingerprint bob
Alice: /verify bob HUELLA_DE_BOB
Bob:   /fingerprint alice
Bob:   /verify alice HUELLA_DE_ALICE
Alice: Hola Bob
Bob:   Hola Alice
```

Comprueba que el otro cliente recibe texto legible y el emisor recibe ACK. Cierra ambos con Ctrl+C, ábrelos con las mismas claves, URL y frases, y comprueba el historial. El ACK indica persistencia, no lectura. `/pending` y `/retry ID` permiten revisar y reintentar envíos sin confirmar con el mismo sobre.

Para validar la LAN, repite el flujo en **dos equipos físicos**. Dos procesos del mismo equipo conectados a su IP privada no comprueban el firewall remoto ni el aislamiento Wi-Fi. Registra sistemas, terminales y resultado sin claves, códigos ni mensajes privados. Consulta las [pruebas y límites actuales](verification.md) y las [operaciones de recuperación y borrado](install.md).

## Detener y conservar datos

Ctrl+C detiene clientes y servidor. `docker compose -f deployments/compose.yml stop db` detiene PostgreSQL si usaste Docker. Conserva base, claves y archivos de estado para volver a usar las cuentas. No uses `down -v` si quieres conservar la base.

El modo E2EE oculta mensajes vencidos a las 720 horas y los purga al arrancar y cada hora. La deduplicación dura 30 días. El cliente solo limpia su archivo al usarse; no borra equipos apagados ni copias externas. Esta guía no configura respaldos automáticos: el administrador debe informar y comprobar la retención de sus copias. Las [herramientas de respaldo](operations.md) se configuran aparte.

## Comprobar el desarrollo

```powershell
$env:TERMLINK_TEST_DATABASE_URL = $env:DATABASE_URL
$env:KARATTY_CRYPTO_BIN = (Resolve-Path bin/karatty-crypto.exe).Path
./scripts/check.ps1
```

Las integraciones crean esquemas aleatorios y los eliminan al terminar; el rol necesita permiso para crearlos. Sin las variables se omiten las pruebas correspondientes. En Linux/macOS exporta las mismas variables y usa los [controles de contribución](project/CONTRIBUTING.md). El alojamiento en Internet queda para la [guía prevista de despliegue](deploy.md).

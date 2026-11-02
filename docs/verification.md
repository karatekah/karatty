# Verificación de la alpha experimental

Esta matriz describe pruebas reproducibles del código para la alpha local/LAN. No sustituye una auditoría independiente ni una prueba entre equipos físicos. VPS, dominio y proveedores quedan para un futuro servicio alojado. Los resultados de cada commit están en [GitHub Actions](https://github.com/karatekah/termlink/actions/workflows/ci.yml).

| Área | Evidencia automatizada |
| --- | --- |
| Registro sin correo y recuperación | `TestAnonymousSignupAndOfflineRecovery`: posesión de clave, desafío de un intento, código incorrecto, revocación y rotación del código. |
| Consentimiento y claves públicas | `TestE2EEKeyDistributionRequiresAcceptedDirect`: no expone identidad ni consume claves antes de aceptación; rechaza firma incorrecta y no reactiva claves consumidas. |
| Mensajes E2EE | `TestE2EEClientExchangeAndLocalPassphrase`: dos clientes, huellas locales coincidentes, envío y respuesta reales, cuerpo opaco en PostgreSQL, 4096 bytes, reinicio, ACK alterado, sobre manipulado y reintento estable. |
| Recuperación del dispositivo | La misma integración reemplaza la clave SSH, rota la identidad Olm, exige nueva verificación, migra el archivo legible y preserva envíos inciertos; comprueba que se elimina el estado anterior. |
| Cierre y borrado | La integración E2EE mantiene la lectura del archivo local de un directo cerrado, impide enviar y retira los archivos de la cuenta borrada. Las suites existentes cubren autorización y borrado del servidor. |
| Secreto futuro acotado | La prueba Rust avanza el ratchet y comprueba que copias de la cuenta y sesión actuales no descifran un mensaje ya consumido. También rechaza repetición y manipulación y recibe mensajes fuera de orden. No protege el archivo local desbloqueado ni una copia antigua del estado. |
| Archivo local | `pkg/localstate`: frase incorrecta, manipulación, bloqueo entre procesos, reemplazo y reapertura de 16 MiB. `pkg/secureclient`: purga a 720 horas y orden del historial al recibir un ACK antes de sincronizar el backlog. |
| Límites de transporte | `TestMaximumEnvelopePageOverHTTPAndWebSocket`: página completa de 100 cuerpos con el tamaño máximo y doble escape JSON, sin truncamiento. |
| TUI | `pkg/tui`: huella propia y del contacto completas, instrucción de comparación externa y recuperación de un envío guardado con el mismo ID. |
| Cliente nativo | `TestE2EEClientNativeTUI` abre dos ejecutables reales sobre PTY, desbloquea sus archivos locales, compara huellas, verifica ambos contactos y entrega un mensaje que PostgreSQL conserva como sobre opaco. |
| TLS local/LAN | `scripts/TestDevelopmentCertificateLAN`: cadena de confianza, localhost, IPv4/IPv6 de loopback e IP privada; rechaza una IP ajena al certificado y sobrescribir certificados. También rechaza IP inválida, indefinida o multicast. |
| Distribución | CI construye, extrae y ejecuta la versión y el puente Go–Rust en Windows, Linux y macOS; comprueba hashes, textos de licencias y commit de origen. |
| Servicio | CI prueba PostgreSQL, TLS, WSS, cliente ejecutable y gateway legado; ejecuta detector de carreras, `go vet`, `govulncheck`, auditoría de Cargo.lock con RustSec, Docker, sintaxis de shell y verificación de unidades systemd. |

## Ensayo de Windows del 1 de noviembre de 2026

`TestE2EEClientExchangeAndLocalPassphrase` pasó con PostgreSQL 17.11 nativo y `karatty-crypto.exe` extraído del paquete Windows del candidato `d39b5da`. La integración comprueba envío, respuesta, historial, recuperación y reintentos con dos clientes; usa TLS en loopback. No es una sesión manual de dos TUI en Windows. La prueba nativa sobre PTY de la tabla se ejecuta en Linux en CI.

También pasaron `go vet ./...`, `go test -count=1 ./...` con base y puente configurados, `go build ./cmd/...`, las pruebas del generador de certificados y la comprobación de enlaces Markdown relativos. Compilar con éxito no implica que el antivirus permita ejecutar el archivo resultante.

En el intento inicial, la compilación del servidor terminó, pero al ejecutarlo en la IP privada Windows Defender bloqueó `.local/lan-alpha/server.exe` con la detección `Behavior:Win32/DefenseEvasion.A!ml`. No se desactivaron protecciones ni se añadió una exclusión. La detección no se ha resuelto ni se ha demostrado que sea un falso positivo. Por ello, el ensayo del servidor nativo Windows por la IP de LAN quedó pendiente; no se afirma que ese despliegue haya pasado.

### Ensayo nativo tras restaurar el archivo

Posteriormente, el usuario restauró `.local/defender-review/server.exe` desde la cuarentena de Windows. Esa compilación del código `4c0356f` tiene SHA-256 `9a81de63cc5149d551bcc4abc3cf0b6a8b83b8658a58d6fc819bb8fe7db830f1`. Después pasaron la ejecución de `--help` y el arranque del servidor nativo en `https://127.0.0.1:28443`, en modo E2EE y sin gateway SSH.

Se extrajo el paquete Windows `v0.1.0-alpha.1` de ese mismo commit y se abrieron dos TUI nativas en terminales ConPTY del mismo equipo. Se probaron registro con claves distintas y frase secreta local, solicitud y aceptación del directo, intercambio de las huellas propias y verificación de ambos contactos. Alice envió un mensaje y Bob respondió; ambos mensajes aparecieron en el otro cliente y recibieron ACK con secuencias 1 y 2. Tras cerrar y volver a abrir las dos TUI, ambas recuperaron los dos mensajes y sincronizaron hasta la secuencia 2.

La consulta a PostgreSQL comprobó dos sobres con estructura E2EE y ausencia de ambos textos de prueba en los cuerpos almacenados. Los textos tampoco aparecieron en los logs operativos del servidor. Al terminar se cerraron los clientes y se detuvo el servidor de ensayo.

La protección en tiempo real y el monitor de comportamiento de Defender permanecieron activos, con firmas `1.463.83.0`. Durante esta prueba no se observó una nueva detección para ese ejecutable. Esto valida esa compilación tras la restauración local en este equipo; no confirma que la clasificación anterior fuera un falso positivo ni garantiza el comportamiento de otras instalaciones de Windows. No se enviaron muestras a Microsoft.

No se dispone de una segunda máquina en esta tarea. La conversación entre **dos equipos físicos de la LAN** sigue sin comprobarse. La [guía](quickstart.md) describe cómo realizarla; las pruebas de certificados no demuestran la conectividad entre máquinas.

## Límites de esta evidencia

- No se ha hecho una auditoría externa de Karatty. La auditoría pública de vodozemac corresponde a esa biblioteca y su versión auditada; no valida automáticamente esta integración ni versiones posteriores.
- Las pruebas comprueban persistencia antes de exponer o enviar mensajes, bloqueos y reemplazo de archivos. No se ha ensayado un corte físico de energía en los sistemas de archivos de las tres plataformas.
- Secuencia, fecha, nombre de cuenta y disponibilidad del servicio son metadatos controlados por el servidor. El cliente valida estructura y vincula criptográficamente conversación, remitente e ID de mensaje dentro del contenido, pero un servidor puede ocultar mensajes o negar servicio.
- El archivo local está acotado a 16 MiB antes de cifrar. Si se agota, el cliente devuelve error y no transmite un ratchet que no haya podido guardar. La limpieza solo puede ejecutarse cuando se usa la aplicación; no promete borrar un equipo apagado ni snapshots ajenos.
- La TUI nativa no tiene aún una revisión manual completa con lectores de pantalla y terminales de todas las plataformas. CI verifica el flujo funcional y los binarios, no esa experiencia visual.
- Faltan la instalación externa contra el servicio real, el vencimiento físico de respaldos remotos, una restauración desde el proveedor contratado, y las verificaciones de dominio, certificado, políticas y contacto del operador. Deben registrarse en [release.md](release.md) antes de anunciar el servicio.

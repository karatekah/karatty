# Verificación de la candidata alpha

Esta matriz describe pruebas reproducibles del código. No sustituye una auditoría independiente ni una prueba del VPS, el dominio o los proveedores que aún no se han contratado. Los resultados de cada commit están en [GitHub Actions](https://github.com/karatekah/termlink/actions/workflows/ci.yml).

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
| Distribución | CI construye, extrae y ejecuta la versión y el puente Go–Rust en Windows, Linux y macOS; comprueba hashes, textos de licencias y commit de origen. |
| Servicio | CI prueba PostgreSQL, TLS, WSS, cliente ejecutable y gateway legado; ejecuta detector de carreras, `go vet`, `govulncheck`, auditoría de Cargo.lock con RustSec, Docker, sintaxis de shell y verificación de unidades systemd. |

## Límites de esta evidencia

- No se ha hecho una auditoría externa de Karatty. La auditoría pública de vodozemac corresponde a esa biblioteca y su versión auditada; no valida automáticamente esta integración ni versiones posteriores.
- Las pruebas comprueban persistencia antes de exponer o enviar mensajes, bloqueos y reemplazo de archivos. No se ha ensayado un corte físico de energía en los sistemas de archivos de las tres plataformas.
- Secuencia, fecha, nombre de cuenta y disponibilidad del servicio son metadatos controlados por el servidor. El cliente valida estructura y vincula criptográficamente conversación, remitente e ID de mensaje dentro del contenido, pero un servidor puede ocultar mensajes o negar servicio.
- El archivo local está acotado a 16 MiB antes de cifrar. Si se agota, el cliente devuelve error y no transmite un ratchet que no haya podido guardar. La limpieza solo puede ejecutarse cuando se usa la aplicación; no promete borrar un equipo apagado ni snapshots ajenos.
- La TUI nativa no tiene aún una revisión manual completa con lectores de pantalla y terminales de todas las plataformas. CI verifica el flujo funcional y los binarios, no esa experiencia visual.
- Faltan la instalación externa contra el servicio real, el vencimiento físico de respaldos remotos, una restauración desde el proveedor contratado, y las verificaciones de dominio, certificado, políticas y contacto del operador. Deben registrarse en [release.md](release.md) antes de anunciar el servicio.

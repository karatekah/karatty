# Instalar y usar Karatty alpha

La primera alpha está orientada a uso local con un servidor administrado por ti. No hay servicio del proyecto en Internet; dominio y VPS quedan para el futuro. Descarga los paquetes de [v0.1.0-alpha.1](https://github.com/karatekah/termlink/releases/tag/v0.1.0-alpha.1). Para iniciar el servidor y probar dos cuentas, sigue la [guía local y LAN](quickstart.md). La LAN entre dos equipos físicos sigue sin comprobarse.

Los comandos de esta página usan un servidor en el mismo equipo. Para LAN, sustituye **en todos los comandos** la URL por la IP privada del servidor y `RUTA_CA` por la ruta a `ca.crt`, recibido y comprobado con su administrador. Usa siempre la misma URL para conservar la identidad local. Los clientes no necesitan PostgreSQL.

## Descargar y verificar

La Release incluye estos paquetes del cliente; el servidor se compila desde el código según la guía local. Otras arquitecturas no tienen descarga precompilada en esta versión.

| Plataforma | Archivo |
| --- | --- |
| Windows x64 | `karatty-v0.1.0-alpha.1-windows-amd64.zip` |
| Linux x64 | `karatty-v0.1.0-alpha.1-linux-amd64.tar.gz` |
| macOS Apple Silicon | `karatty-v0.1.0-alpha.1-darwin-arm64.tar.gz` |

Descarga también `SHA256SUMS` de la misma versión y compara el hash antes de extraer:

```powershell
# Windows: compara el resultado con la línea del archivo descargado.
Get-FileHash .\karatty-v0.1.0-alpha.1-windows-amd64.zip -Algorithm SHA256
```

```sh
# Linux
sha256sum -c SHA256SUMS --ignore-missing
# macOS: compara el resultado con SHA256SUMS.
shasum -a 256 karatty-v0.1.0-alpha.1-darwin-arm64.tar.gz
```

Extrae el paquete en una carpeta propia y conserva juntos `karatty` y `karatty-crypto` (con `.exe` en Windows). Puedes añadir esa carpeta a tu PATH. La descarga incluye sus licencias, las dependencias fijadas y `BUILD_INFO` con el commit de origen. Los paquetes alpha no están firmados ni notarizados; los hashes comprueban integridad respecto del manifiesto publicado, no autentican por sí solos al autor.

Windows Defender bloqueó una compilación del servidor durante el ensayo. Su propietario la restauró y después pasaron las pruebas locales con las protecciones activas. La clasificación no se ha confirmado errónea. Si tu antivirus bloquea un archivo, conserva la alerta y revisa su procedencia; no desactives la protección para instalar. El [registro del ensayo](verification.md) identifica la compilación probada y sus límites.

## Crear una identidad

Usa una clave SSH Ed25519 dedicada a Karatty. OpenSSH pide una frase para protegerla; elige también una frase independiente para el archivo local de mensajes.

```sh
ssh-keygen -t ed25519 -f RUTA_DE_TU_CLAVE
karatty register --server https://127.0.0.1:8443 --ca-cert RUTA_CA --key RUTA_DE_TU_CLAVE --username TU_SEUDONIMO --public
```

En PowerShell, usa `./karatty.exe` si ejecutas desde la carpeta extraída. El seudónimo admite letras ASCII, números, guion y guion bajo, hasta 64 caracteres. No se pide correo ni nombre real para crear la cuenta. El servidor registra tu clave pública y metadatos de cuenta y conversación; un seudónimo no garantiza anonimato absoluto.

Guarda el código de recuperación que aparece una sola vez **fuera de este equipo**. Después se solicita y confirma la frase secreta local de al menos 12 caracteres. Esa frase no se envía al servidor. El archivo local solo puede abrirse con ella; el código de recuperación no la sustituye.

## Abrir un directo

```sh
karatty chat --server https://127.0.0.1:8443 --ca-cert RUTA_CA --key RUTA_DE_TU_CLAVE
```

Introduce la frase de la clave SSH si está protegida y la del archivo local. Dentro de la TUI:

1. Alice escribe `/direct bob` para solicitar el chat.
2. Bob usa `/requests` y acepta con `/direct alice`.
3. Ambos usan `/myfingerprint` para obtener su propia huella y la comparten por otro medio que ya confíen.
4. Alice consulta `/fingerprint bob`, compara el valor con el que Bob le dio y escribe `/verify bob HUELLA_DE_BOB`. Bob hace lo mismo con la de Alice.
5. Escribe un mensaje y pulsa Enter. Tab cambia de chat; PgUp/PgDn recorren el historial; Ctrl+L recarga; Ctrl+C sale.

La verificación por otro medio permite detectar que un servidor sustituya una identidad en el primer contacto. Una huella copiada únicamente de la misma conversación no ofrece esa comprobación independiente. La TUI detiene el envío si cambia la identidad del contacto.

El ACK confirma que el servidor guardó el sobre cifrado; no confirma que la otra persona lo haya leído. Ctrl+R reintenta un envío sin confirmar con el mismo ID y sobre. Tras reiniciar, `/pending` lista los envíos guardados del directo y `/retry ID` reintenta uno. En CLI, `pending` los lista y `retry-pending --conversation ID_DEL_CHAT --id ID_DEL_ENVIO` lo reintenta. Esc permite editar la entrada y conserva el envío guardado para revisión; abandonar un envío incierto no puede retirar una copia que el servidor ya haya aceptado. `/decline USUARIO` rechaza solicitudes, `/block USUARIO` cierra el directo y `/unblock USUARIO` permite recibir nuevas solicitudes.

## Historial, recuperación y borrado

Se conservan mensajes durante 720 horas (30 días). El servidor oculta los vencidos y los purga al arrancar y cada hora. El cliente limpia su archivo local al abrirlo o guardarlo y oculta entradas vencidas al leerlo; un equipo apagado no puede ejecutar esa limpieza. La guía local no configura respaldos automáticos. Si el administrador hace copias, debe informar y comprobar su retención. La ventana máxima de siete días adicionales es un requisito del futuro servicio alojado, todavía sin operar. El secreto futuro no protege una copia local ya desbloqueada ni un equipo comprometido mientras se usa.

Si pierdes la clave SSH, genera otra y ejecuta:

```sh
karatty recover --server https://127.0.0.1:8443 --ca-cert RUTA_CA --key NUEVA_CLAVE --username TU_SEUDONIMO
```

Introduce el código guardado y conserva el código nuevo. Luego abre el cliente con la clave nueva. Si el archivo anterior sigue en este equipo, cierra otros clientes y usa la misma frase secreta local para migrar el historial ya legible. El estado anterior se retira después de guardar el nuevo. Los envíos anteriores sin ACK se listan con `karatty uncertain` y no se reenvían solos. Si perdiste el archivo o su frase, recuperar la cuenta no restaura esos mensajes.

El contacto debe comparar tu nueva huella por otro medio y usar `/reverify USUARIO HUELLA`. Si conserva envíos pendientes, `/discardpending USUARIO` permite abandonarlos explícitamente antes de aceptar la nueva identidad. Comprueba si llegaron antes de repetirlos. Los mensajes antiguos sin descifrar pueden figurar como no disponibles.

Para borrar la cuenta:

```sh
karatty delete-account --server https://127.0.0.1:8443 --ca-cert RUTA_CA --key RUTA_DE_TU_CLAVE --confirm-delete
```

Se borran tu cuenta, claves, sesiones y mensajes propios de la base activa y los archivos locales de esa cuenta en el directorio configurado de este equipo. Cierra otros clientes para que la limpieza local pueda adquirir el bloqueo. Las copias de otras personas y otros equipos no se pueden retirar a distancia; los respaldos siguen su vencimiento. Si el servidor borró la cuenta pero falló la limpieza local, el CLI muestra la ruta que requiere revisión.

## Actualizar o desinstalar

Para actualizar, cierra Karatty, verifica y extrae la nueva descarga, y reemplaza ambos ejecutables juntos. Conserva tu clave SSH y el directorio de estado. La alpha admite un solo dispositivo por cuenta.

El estado se guarda por defecto en la carpeta de configuración del usuario, bajo `karatty`: `%AppData%\karatty` en Windows, `$XDG_CONFIG_HOME/karatty` o `~/.config/karatty` en Linux, y `~/Library/Application Support/karatty` en macOS. `--state-dir RUTA_ABSOLUTA` permite elegir otro directorio. Desinstalar los ejecutables no borra una cuenta del servidor; usa `delete-account` antes si también quieres eliminarla.

Arquitectura y tecnologías: [architecture.md](architecture.md). Garantías y límites del cifrado: [ADR 0003](adr/0003-e2ee-1to1.md). Desarrollo y pruebas: [CONTRIBUTING](../CONTRIBUTING.md).

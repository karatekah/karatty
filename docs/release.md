# Preparación de la primera alpha pública

Esta es una lista de comprobación, no un anuncio de disponibilidad. El repositorio sigue privado. El modo público de ensayo ya usa el cliente E2EE local; el modo legado por invitación sigue enviando texto legible al servidor. La publicación incluirá **el código fuente de este proyecto**, documentación y los dos ejecutables locales en una versión identificable.

## Estado de la candidata

| Parte | Estado |
| --- | --- |
| Cliente, E2EE, recuperación, reintentos y retención local | Implementados; evidencia en [verification.md](verification.md). |
| Paquetes nativos y licencias | CI construye, extrae, verifica y guarda descargas de Windows, Linux y macOS. |
| Instalación y documentación | [Primer uso](install.md), arquitectura, protocolo, seguridad, roadmap y changelog preparados. |
| Instalación del servidor | [Guía](deploy.md), listener TLS explícito, unidades systemd, hook de certificado y copias preparados; no instalados en un VPS. |
| Dominio y proveedores | `karatty.dev` elegido; compra, VPS, región y bucket pendientes. |
| Privacidad y condiciones | Borradores; faltan identidad y correo público del operador y revisión aplicable. |
| Pruebas remotas | Pendientes de disponer de la infraestructura contratada. |
| GitHub público | Repositorio privado; preparar un borrador de Release no cambia su visibilidad ni publica el servicio. |

La candidata técnica puede revisarse y probarse localmente. La activación pública requiere los datos y recursos de las filas pendientes; no deben completarse con valores inventados.

## Condiciones antes de abrir el repositorio

1. Completar E2EE de directos 1:1 con secreto futuro según el [ADR 0003](adr/0003-e2ee-1to1.md) en el cliente TUI local. Probar envío, recepción, identidad del contacto, manipulación de sobres, pérdida y cambio de claves, reintentos, historial y reconexión. Confirmar que la API, PostgreSQL y los logs no reciben texto legible. Retirar canales y TUI SSH remota de la configuración pública.
2. Ejecutar las pruebas y compilaciones de CI en el commit candidato. Ensayar de forma real el registro seudónimo, recuperación, consentimiento, bloqueo, borrado de cuenta y retención de 30 días.
3. Contratar y comprobar el dominio y la infraestructura; ensayar copia cifrada, restauración y eliminación remota de respaldos. Completar política de privacidad y condiciones con la identidad y el contacto público del operador. La [guía de operación](operations.md) enumera estas tareas.
4. Verificar que el archivo `LICENSE` de Apache-2.0 y los avisos de terceros cubren la distribución. Revisar `SECURITY.md` y habilitar el reporte privado de GitHub al abrir el repositorio; revisar `CONTRIBUTING.md` con el flujo de desarrollo y pruebas. Evitar publicar una dirección que aún no existe.
5. Revisar **todo el historial Git**, las identidades y correos de autor de commits, los archivos actuales y los registros de GitHub Actions para secretos, datos personales, claves, volcados, certificados privados o material que no deba hacerse público. Al volver público un repositorio, GitHub también hace visibles el historial y los registros de Actions. Comprobar que `.local/`, `.env`, `bin/` y `dist/` siguen ignorados. Si aparece información sensible, resolverla antes de cambiar la visibilidad.
6. Revisar el README y las guías desde una instalación limpia en Windows, Linux y macOS. El README deberá explicar qué hace la alpha, qué limita, cómo instalarla, un ejemplo de conversación 1:1, cómo verificar la descarga y dónde encontrar arquitectura, seguridad, privacidad, contribuciones y roadmap. Describir únicamente funciones comprobadas.
7. Preparar la etiqueta, notas de cambios, compatibilidad de protocolo y un borrador de GitHub Release para el mismo commit probado. Tras cambiar la visibilidad, habilitar el reporte privado de vulnerabilidades de GitHub, publicar la Release y confirmar que el repositorio y los archivos son accesibles sin iniciar sesión antes de anunciar la alpha.

El cambio de visibilidad se hará una vez cumplidas y registradas las condiciones anteriores; después se publicará y comprobará la Release. No se ha realizado ese cambio.

## Construir candidatos del cliente

Desde PowerShell, con Go 1.27.3 y Rust 1.99.1 disponibles en la plataforma de destino:

```powershell
./scripts/build-release.ps1 -Version v0.1.0-alpha.1
```

El script compila para la plataforma nativa `karatty` sin CGo y `karatty-crypto` con Rust. Crea una carpeta y archivo comprimido por plataforma en `dist/VERSIÓN/`, con hashes de ambos ejecutables y del archivo. Incluye licencia propia, textos de terceros, locks de dependencias y `BUILD_INFO` con el commit. Se niega a sobrescribir una carpeta existente. No incluye el servidor, claves, configuración ni base de datos. CI construye, extrae y ejecuta el puente de cada paquete y conserva las descargas como artifacts de Actions para Windows, Linux y macOS.

Antes de adjuntar archivos, comparar el commit con un CI verde, ejecutar `karatty version` y comprobar el hash del archivo y de ambos ejecutables. Instalar cada paquete descargado en otra máquina de su plataforma y probar registro, comparación de huellas, chat y actualización de versión contra el servicio publicado. En Windows, `Get-FileHash ARCHIVO -Algorithm SHA256`; en Linux/macOS, `sha256sum -c SHA256SUMS`.

La primera alpha no tendrá firma de código ni notarización para macOS. Un hash publicado junto al binario detecta cambios frente a ese archivo, pero no autentica por sí solo al publicador. Explicar esta limitación en las instrucciones de descarga. Como el cliente apunta hoy a localhost por defecto, documentar `--server` con la dirección real antes de distribuirlo.

## Artefactos de la Release

- Etiqueta y notas de versión enlazadas al commit verificado.
- Paquetes nativos que incluyan `karatty`, `karatty-crypto` y sus hashes, más el manifiesto de hashes de los archivos comprimidos.
- Instrucciones de instalación, primer uso, actualización y eliminación.
- Enlaces a la documentación de arquitectura, protocolo, seguridad, privacidad, límites y roadmap.
- Resultado fechado de las pruebas operativas y limitaciones conocidas de la alpha.

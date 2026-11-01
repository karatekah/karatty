# Preparación de la primera alpha pública

Esta es una lista de comprobación, no un anuncio de disponibilidad. El repositorio sigue privado y el servidor todavía puede leer mensajes. La publicación incluirá **el código fuente de este proyecto**, documentación y ejecutables del cliente en una versión identificable. No se creará un repositorio de solo binarios como sustituto del código público.

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

Desde PowerShell, con Go 1.27.3 disponible:

```powershell
./scripts/build-release.ps1 -Version v0.1.0-alpha.1
```

El script compila `karatty` sin CGo para Windows amd64, Linux amd64/arm64 y macOS amd64/arm64 en `dist/VERSIÓN/`. Escribe `SHA256SUMS` con un hash por ejecutable y se niega a sobrescribir una carpeta existente. No incluye el servidor, claves, configuración ni base de datos. Si el diseño final de E2EE cambia las dependencias de compilación, actualizar el script y CI antes de la Release.

Antes de adjuntar archivos, comparar el commit con un CI verde, ejecutar `karatty version` en Windows y comprobar los cinco hashes. Instalar al menos un binario descargado desde la Release en otra máquina y probar registro, inicio de sesión, chat y actualización de versión contra el servicio publicado. En Windows, `Get-FileHash ARCHIVO -Algorithm SHA256`; en Linux/macOS, `sha256sum -c SHA256SUMS`.

La primera alpha no tendrá firma de código ni notarización para macOS. Un hash publicado junto al binario detecta cambios frente a ese archivo, pero no autentica por sí solo al publicador. Explicar esta limitación en las instrucciones de descarga. Como el cliente apunta hoy a localhost por defecto, documentar `--server` con la dirección real antes de distribuirlo.

## Artefactos de la Release

- Etiqueta y notas de versión enlazadas al commit verificado.
- Cinco clientes para las plataformas indicadas y un archivo `SHA256SUMS`.
- Instrucciones de instalación, primer uso, actualización y eliminación.
- Enlaces a la documentación de arquitectura, protocolo, seguridad, privacidad, límites y roadmap.
- Resultado fechado de las pruebas operativas y limitaciones conocidas de la alpha.

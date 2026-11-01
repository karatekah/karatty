# Preparación de ejecutables públicos

El código fuente permanece en este repositorio privado. El canal público de descarga todavía no existe; estos pasos producen candidatos, no una publicación.

## Construir

Desde PowerShell, con Go 1.27.3 disponible:

```powershell
./scripts/build-release.ps1 -Version v0.1.0-alpha.1
```

El script compila el cliente `karatty` sin CGo para Windows amd64, Linux amd64/arm64 y macOS amd64/arm64 en `dist/VERSIÓN/`. También escribe `SHA256SUMS` con un hash por ejecutable. Se niega a sobrescribir una carpeta de salida existente. No incluye el servidor, claves, configuración, base de datos ni código fuente.

Antes de publicar, compara el commit de compilación con un CI verde, ejecuta el cliente Windows con `version` y comprueba todos los hashes. Una vez activo el servicio, instala al menos un binario desde la ubicación pública en una máquina distinta y prueba registro, login, chat y actualización de versión. En Windows, `Get-FileHash ARCHIVO -Algorithm SHA256` permite comparar el hash; en Linux/macOS, `sha256sum -c SHA256SUMS`.

La alpha no tiene firma de código ni notarización para macOS. Los hashes comprueban integridad respecto del archivo `SHA256SUMS` publicado; no sustituyen una firma verificable. Indica esta limitación en la página de descarga. No publiques binarios que aún apunten solo al servidor local sin instrucciones claras para `--server`.

## Publicar después de las puertas de lanzamiento

Crear un canal público separado, por ejemplo un repositorio de distribución de `karatekah` que contenga solo notas de versión, instrucciones y binarios, o una página de descargas en el dominio elegido. Verificar que el repositorio privado no se adjunte ni se vuelva público. Adjuntar los cinco clientes y `SHA256SUMS` con la misma versión, documentar los cambios y comprobar la descarga anónima. No se ha creado todavía ese canal ni comprado el dominio.

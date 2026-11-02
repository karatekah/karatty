# Publicación de Karatty v0.1.0-alpha.1

La [primera alpha](https://github.com/karatekah/termlink/releases/tag/v0.1.0-alpha.1) distribuye código fuente, documentación y paquetes del cliente para uso local con servidor propio. Es experimental. La configuración de LAN está documentada, pero la conversación entre dos equipos físicos aún no se ha comprobado. El proyecto no ofrece alojamiento en Internet.

## Alcance y evidencia

| Parte | Estado |
| --- | --- |
| Cliente, E2EE, recuperación, reintentos y retención | Implementados; pruebas y límites en [verification.md](verification.md). |
| Ensayo Windows | Servidor TLS y dos TUI nativas: registro, aceptación, huellas, mensajes y reapertura del historial. |
| Defender | La compilación del ensayo necesitó restaurarse desde cuarentena. Defender permaneció activo; no se ha confirmado un falso positivo ni validado otras instalaciones. |
| Paquetes | Windows x64, Linux x64 y macOS Apple Silicon, con ambos ejecutables, licencias, locks, BUILD_INFO y hashes. El servidor se compila desde el código. |
| LAN entre dos equipos físicos | Pendiente. Se publica como configuración experimental, sin afirmar que esa prueba pasó. |
| Dominio, VPS y proveedores | Trabajo futuro; no se necesitan para esta alpha local. |
| Privacidad y condiciones del futuro alojamiento | Borradores sin vigencia. Cada administrador debe informar sobre su instancia y copias. |
| Reportes de seguridad | Canal privado de GitHub según [SECURITY.md](../SECURITY.md). |

El código propio usa Apache-2.0; las dependencias conservan sus licencias. Las descargas no están firmadas ni notarizadas. No hay auditoría independiente de Karatty.

## Comprobaciones de publicación

1. Registrar el resultado de CI para el commit que se etiqueta: PostgreSQL real, E2EE, TUI, Go/Rust, carreras, extracción de paquetes y avisos de dependencias. La prueba local Windows corresponde a la compilación identificada en la matriz; no se extiende automáticamente a otros archivos o equipos.
2. Revisar el historial completo y los archivos actuales para secretos, datos personales, claves, volcados o configuración privada. Revisar también los logs retenidos de Actions, que pasan a ser públicos con el repositorio. Mantener `.local/`, `.env`, `bin/` y `dist/` ignorados.
3. Revisar README, instalación, arquitectura, seguridad, contribuciones y changelog. Mantener visibles las limitaciones de Defender y LAN; los borradores de un futuro servicio alojado no son políticas activas.
4. Integrar la rama de preparación en `main`, conservar los commits convencionales y verificar CI en el commit final. Descargar sus paquetes y comprobar SHA-256, licencias y el commit de `BUILD_INFO`.
5. Crear la etiqueta `v0.1.0-alpha.1` sobre ese commit, adjuntar los tres paquetes y un manifiesto `SHA256SUMS` conjunto, y preparar las notas con los resultados y límites.
6. Hacer público el repositorio, habilitar el reporte privado de vulnerabilidades y publicar la Release como **prerelease**. Verificar sin autenticación el repositorio, README, etiqueta, página de Release y descargas.

El ensayo entre dos equipos físicos será necesario antes de declarar validada la LAN. La compra de dominio, VPS, buzones y las pruebas de proveedores pertenecen al [futuro servicio alojado](operations.md).

## Construir paquetes del cliente

Desde PowerShell, con Go 1.27.3 y Rust 1.99.1 en la plataforma de destino:

```powershell
./scripts/build-release.ps1 -Version v0.1.0-alpha.1
```

El script construye para la plataforma nativa `karatty` sin CGo y `karatty-crypto` con Rust. Crea una carpeta y archivo comprimido en `dist/VERSIÓN/`, con hashes de ambos ejecutables y del archivo. Incluye licencia propia, textos de terceros, locks de dependencias y `BUILD_INFO` con el commit. Rechaza sobrescribir una carpeta existente. No incluye claves, configuración ni base de datos. El servidor se construye por separado según [quickstart.md](quickstart.md).

CI construye, extrae y ejecuta el puente de cada paquete en Windows, Linux y macOS. Sus descargas de Actions son insumos de la Release; no sustituyen su publicación y verificación. La [guía de instalación](install.md) explica cómo comprobar los hashes, conservar ambos ejecutables juntos, crear la identidad y conversar.

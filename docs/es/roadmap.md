# Alcance y decisiones de producto

[English](../roadmap.md) · **Español** · [Italiano](../it/roadmap.md)

La primera alpha es software para usar en un equipo o una red local: servidor propio con PostgreSQL, clientes TUI locales y directos 1:1 con aceptación. No requiere contratar dominio, VPS ni correo. El alojamiento de un servicio del proyecto en Internet queda para una fase futura; `karatty.dev` es el dominio elegido, todavía sin comprar.

## Alpha local y LAN

- Identidad seudónima con clave SSH Ed25519, sin correo ni nombre real. Recuperación con código guardado fuera del equipo.
- E2EE con secreto futuro acotado según el [ADR 0003](adr/0003-e2ee-1to1.md); comparación de huellas por otro medio y nueva verificación al cambiar la identidad.
- Archivo local protegido con frase secreta, reintentos persistidos y un dispositivo por cuenta.
- Retención de 720 horas: se ocultan y purgan cuerpos e identificadores vencidos; la deduplicación se garantiza solo durante esos 30 días. La purga se activa automáticamente en `--public-signup`. Las copias externas y equipos apagados tienen límites distintos.
- Código, documentación y paquetes del cliente distribuidos en GitHub bajo Apache-2.0. El [registro de publicación](release.md) describe el proceso y los límites de esta alpha local.
- Pruebas reproducibles y límites en [verification.md](verification.md), incluido el ensayo de dos TUI Windows tras restaurar el ejecutable desde la cuarentena de Defender. La LAN entre dos equipos físicos y el comportamiento de la detección en otras instalaciones siguen sin comprobarse.

## Servicio alojado futuro

La compra del dominio, VPS, bucket, buzones, políticas con identidad del operador y pruebas de respaldo/restauración remota pertenecen a este hito. No bloquean por sí mismas la distribución de la alpha local/LAN. Antes de ofrecer alojamiento, deben cumplirse las condiciones de [operación](operations.md) y [despliegue](deploy.md). La ventana máxima elegida para respaldos es de siete días adicionales; debe comprobarse en la infraestructura real antes de prometerla.

## Después de la alpha

Grupos, SSH remoto con un modelo de cifrado adecuado, múltiples dispositivos, adjuntos, búsqueda, notificaciones e integraciones requieren diseño y pruebas propios. No tienen una fecha comprometida. Los posibles planes Community/Pro/Teams siguen siendo hipótesis.

Karatty es el nombre de trabajo elegido por el usuario. La búsqueda preliminar registró un juego [Karatty VR](https://usb50-zx-spectrum.itch.io/ka). RDAP respondió como no registrado el 31 de octubre de 2026; esto no reserva el dominio ni garantiza su compra. El [buscador de Indecopi](https://pi.indecopi.gob.pe/buscatumarca/#/inicio) mostró cero marcas vigentes en búsqueda precisa para «Karatty» en clases 9, 38 y 42 el 1 de noviembre de 2026; la búsqueda amplia devolvió 200 resultados. La revisión de similitudes sigue pendiente. El módulo Go y el repositorio actuales son `karatekah/karatty`. Se conservan identificadores compatibles como `cmd/termlink` y `TERMLINK_*`; el código de la etiqueta histórica v0.1.0-alpha.1 mantiene su módulo original.

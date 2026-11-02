# Idiomas de la documentación

[English](../translations.md) · **Español** · [Italiano](../it/translations.md)

El inglés es el idioma principal en las rutas normales: `README.md`, políticas de la raíz y guías de `docs/`. Las versiones completas en español e italiano están en `docs/es/` y `docs/it/`. Cada página enlaza directamente con la misma página en los otros idiomas.

## Estructura

| Fuente inglesa | Español | Italiano |
| --- | --- | --- |
| `README.md` | `docs/es/README.md` | `docs/it/README.md` |
| `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md` | Mismo nombre en `docs/es/project/` | Mismo nombre en `docs/it/project/` |
| `docs/install.md` y otras guías | `docs/es/install.md` y equivalentes | `docs/it/install.md` y equivalentes |
| `docs/adr/0003-e2ee-1to1.md` | `docs/es/adr/0003-e2ee-1to1.md` | `docs/it/adr/0003-e2ee-1to1.md` |
| `crypto/README.md`, `third_party/README.md` | Subdirectorio equivalente en `docs/es/` | Subdirectorio equivalente en `docs/it/` |

El catálogo es `docs/languages.json`: incluye notas de versión, operación y borradores legales no vigentes, además de los README. Los textos originales de licencias y copyright son los autorizados y no se traducen.

## Actualizar una página

1. Actualiza el inglés y ambas traducciones completas en el mismo cambio. Conserva las salvedades de seguridad, el estado experimental, las cifras, versiones y fechas de evidencia.
2. Mantén literales los ejecutables, flags, variables, rutas API, campos JSON y dominios firmados. Traduce explicaciones, títulos, comentarios y marcadores claramente ilustrativos.
3. Usa enlaces relativos que conserven el idioma elegido. Los enlaces a código, configuración y licencias vuelven al archivo real del repositorio.
4. Conserva la barra de tres idiomas después del título, con el idioma actual marcado.
5. Ejecuta `./scripts/check-docs.ps1` desde PowerShell. Comprueba catálogo, tres versiones, navegación y destinos relativos. Revisa también el Markdown renderizado, las tablas y los bloques de código.

El verificador no demuestra equivalencia semántica: el revisor debe comparar el significado técnico. No sustituyas traducciones por resúmenes ni por inglés de forma silenciosa. Este cambio documental no traduce los mensajes de la TUI en ejecución.

## Nombre y versiones

El repositorio y módulo canónicos son `karatekah/karatty`. No reutilices el nombre anterior del repositorio, pues GitHub mantiene sus redirecciones. Las rutas y variables compatibles conservan su escritura documentada.

La etiqueta, binarios y checksums de `v0.1.0-alpha.1` permanecen intactos. El código y documentación archivados usan el módulo e idioma anteriores. Las guías actuales y los enlaces de idioma de la Release llevan a documentación mantenida; no reescriben ese código histórico.

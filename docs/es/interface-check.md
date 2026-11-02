# Verificación de la interfaz del sprint

[English](../interface-check.md) · **Español** · [Italiano](../it/interface-check.md)

Aplicación de `make-interfaces-feel-better`, alcance: TUI compartida Bubble Tea v2/Bubbles v2. Estilo: texto monoespaciado, separadores y estados escritos; sin sistema CSS ni animaciones propias. No es una revisión visual de Windows Terminal/macOS/Linux completa.

| Categoría | Evidencia inspeccionada | Resultado |
| --- | --- | --- |
| Tipografía | TUI usa ajuste al ancho con ansi.Wrap/Truncate y ancho del terminal | Ajuste implementado; revisión visual manual pendiente |
| Superficies | Cabecera, pestañas, historial y entrada delimitados por texto | Sin dependencias visuales adicionales |
| Animación | Sin animación personalizada; solo cursor estándar | No aplica reproducción a 10% |
| Iconos | Estados de conexión/ACK escritos y selección entre corchetes | No dependen de color ni iconos |
| Rendimiento | Máximo 500 mensajes y página de 100 | Memoria de historial acotada; aún sin perfil de rendimiento bajo carga |

| Severidad | Ubicación | Antes | Después | Motivo |
| --- | --- | --- | --- | --- |
| Media | pkg/tui/model.go, envío y recarga | Recargar podía descartar el ID pendiente | La recarga espera confirmar o cancelar el envío | Evita reintentos ambiguos |
| Media | pkg/tui/model.go, entrada | Entrada fuera del límite no daba explicación | Muestra el máximo de bytes y restricción de controles | Hace visible el siguiente paso |

Candidatos descartados: colores o efectos para conexión (texto funciona en terminales diversos); animación en mensajes frecuentes (añade distracción); nuevos estilos de componentes (los componentes Bubbles existentes bastan para la primera prueba).

Verificación ejecutada: PTY SSH real abre la misma TUI y renderiza cabecera y mensajes persistidos; pruebas de transporte validan secuencias de control. No verificado visualmente: lectores de pantalla, zoom, terminal de 12 columnas y terminales nativos de las tres plataformas. La revisión manual de accesibilidad sigue pendiente y se documenta como límite de la alpha experimental. La prueba funcional del sprint cubre el flujo principal del prototipo; el ensayo E2EE nativo posterior está registrado en verification.md.

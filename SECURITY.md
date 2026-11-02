# Reportar una vulnerabilidad

La versión soportada es **v0.1.0-alpha.1**, experimental. Usa E2EE en directos 1:1 desde la TUI local y tiene [pruebas documentadas](docs/verification.md); no cuenta con una auditoría independiente. El modo legado por invitaciones y la TUI ejecutada por SSH remoto envían texto legible al servidor y quedan fuera de la alpha.

Usa **[Security → Report a vulnerability](https://github.com/karatekah/termlink/security/advisories/new)** en GitHub para enviar un informe privado. No publiques detalles explotables, claves, tokens ni conversaciones en un issue o una discusión. Si el botón no aparece, abre un issue sin detalles técnicos para avisar que falta el canal privado; el mantenedor habilitará ese canal antes de pedir información sensible.

Incluye la versión o commit, plataforma, pasos de reproducción y el impacto observado. Usa datos de prueba. El mantenedor acusará recibo y coordinará la corrección y divulgación con la persona informante. Durante la alpha no se promete un plazo fijo de respuesta.

No hay versiones estables ni versiones públicas anteriores que reciban parches. Durante la alpha, aplica las actualizaciones de seguridad que se publiquen y consulta sus notas de compatibilidad. El proyecto distribuye software; no ofrece todavía un servicio alojado ni un correo público de soporte.

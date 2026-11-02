# Reportar una vulnerabilidad

Karatty aún no tiene una versión pública. La candidata usa E2EE en directos 1:1 desde la TUI local y tiene [pruebas documentadas](docs/verification.md); no cuenta con una auditoría independiente. El modo legado por invitaciones y la TUI ejecutada por SSH remoto envían texto legible al servidor y quedan fuera de la alpha pública.

Cuando el repositorio sea público, usa **Security → Report a vulnerability** en GitHub para enviar un informe privado. No publiques detalles explotables, claves, tokens ni conversaciones en un issue o una discusión. Si el botón no aparece después del lanzamiento, abre un issue sin detalles técnicos para avisar que falta el canal privado; el mantenedor habilitará ese canal antes de pedir información sensible.

Incluye la versión o commit, plataforma, pasos de reproducción y el impacto observado. Usa datos de prueba. El mantenedor acusará recibo y coordinará la corrección y divulgación con la persona informante. Durante la alpha no se promete un plazo fijo de respuesta.

No hay versiones públicas anteriores que reciban parches de seguridad. Esta política se actualizará con las versiones soportadas y el contacto público del operador antes del lanzamiento.

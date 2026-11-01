# Puente criptográfico experimental

Este ejecutable de prueba usa vodozemac 0.11.1 para crear identidades Olm, cifrar y descifrar mensajes. Acepta un objeto JSON por stdin y devuelve uno por stdout. Las pruebas demuestran intercambio en ambas direcciones, rechazo de alteración y rechazo de repetición.

**Aún no forma parte de `karatty` ni del protocolo de red.** Los campos `account` y `session` contienen claves privadas sin cifrar mientras viajan por los pipes locales. No deben guardarse, imprimirse en logs ni enviarse al servidor. Este ejecutable tampoco verifica la identidad del contacto: el cliente deberá comprobar firmas, fijar la huella y detenerse ante cambios de clave. El cliente final deberá cifrar su estado local y guardar el ratchet avanzado y el mensaje pendiente de forma atómica antes de enviar.

Para probarlo desde `crypto/`: `cargo test --locked`. El [ADR 0003](../docs/adr/0003-e2ee-1to1.md) describe la integración y las pruebas que faltan antes de publicar una alpha E2EE.

# Puente Olm local

[English](../../../crypto/README.md) · **Español** · [Italiano](../../it/crypto/README.md)

El ejecutable `karatty-crypto` forma parte del cliente alpha local. Usa vodozemac 0.11.2 para crear identidades Olm, reponer claves de un solo uso y cifrar/descifrar mensajes. Acepta un objeto JSON por stdin y devuelve uno por stdout. Las pruebas Rust cubren intercambio bidireccional, alteración, repetición, entrega fuera de orden y secreto futuro acotado. `pkg/cryptobridge` prueba la invocación desde Go y las firmas de claves iniciales y repuestas.

Los campos `account` y `session` contienen material de claves privadas sin cifrar mientras atraviesan **pipes de procesos locales**. Nunca los registres en logs, los guardes sin cifrar o los envíes al servidor. El puente no verifica la identidad del contacto. `pkg/secureclient` verifica los paquetes de claves firmadas y las huellas, se detiene ante cambios de identidad y persiste atómicamente el ratchet avanzado y el sobre pendiente antes de transmitir. `pkg/localstate` protege el archivo con frase secreta local, Argon2id y XChaCha20-Poly1305.

Ejecuta `cargo test --locked` desde `crypto/`. Compila con `cargo build --locked --release` y conserva el ejecutable resultante junto a `karatty`. Consulta [ADR 0003](../adr/0003-e2ee-1to1.md) y la [verificación](../verification.md) para conocer la evidencia y sus límites. Karatty no tiene auditoría independiente.

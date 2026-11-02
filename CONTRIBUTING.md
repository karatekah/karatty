# Contribuir a Karatty

El repositorio sigue privado mientras se prepara la primera alpha. Esta guía describe cómo reproducir el entorno y revisar cambios; se verificará de nuevo antes de aceptar contribuciones públicas.

## Preparar el entorno

Usa Go 1.27.3, Rust 1.99.1 y PostgreSQL 17.11. La [guía local](docs/quickstart.md) muestra cómo iniciar PostgreSQL con Docker, generar certificados de desarrollo y ejecutar el cliente. Las pruebas de integración requieren `TERMLINK_TEST_DATABASE_URL` apuntando a una **base de desarrollo** con permiso para crear esquemas. `KARATTY_CRYPTO_BIN` debe apuntar al puente compilado para ejecutar la integración E2EE. Nunca apuntes las pruebas a producción.

```sh
go mod verify
cargo test --locked --manifest-path crypto/Cargo.toml
cargo build --locked --release --manifest-path crypto/Cargo.toml
# Define KARATTY_CRYPTO_BIN con la ruta absoluta a crypto/target/release/karatty-crypto.
go vet ./...
go test -count=1 ./...
go build ./cmd/...
```

En PowerShell puedes ejecutar `./scripts/check.ps1`. CI también ejecuta las pruebas con detector de carreras en Linux, compila en Windows/Linux/macOS y prueba el arranque con Docker. Si no se define la variable de base, las pruebas de integración se omiten; una ejecución local verde en ese caso no las sustituye.

## Proponer cambios

Abre un issue para cambios de comportamiento o de protocolo antes de hacerlos. Mantén cada cambio enfocado y explica qué problema resuelve, qué archivos modifica, cómo se probó y qué límites quedan. Los commits usan prefijos convencionales como `feat:`, `fix:`, `docs:`, `test:` o `chore:`. Las migraciones deben preservar los datos existentes o documentar claramente cualquier pérdida inevitable.

El E2EE, la identidad, la retención y el borrado afectan la seguridad y privacidad. Para cambios en esas áreas, documenta el modelo de amenazas, compatibilidad de protocolo y pruebas de fallos. No introduzcas primitivas criptográficas propias ni afirmes garantías que las pruebas no demuestran.

## Reportes de seguridad

No publiques vulnerabilidades explotables ni material sensible en issues. Antes de abrir el repositorio se configurará una vía privada de reporte en `SECURITY.md`. Hasta que esa vía exista y se haya probado, no se abrirá el proyecto a contribuciones públicas.

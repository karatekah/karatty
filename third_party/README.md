# Textos de licencia complementarios

Los paquetes de `matrix-pickle` y `matrix-pickle-derive` 0.2.3 no incluyen el archivo de licencia del workspace en los crates publicados. `matrix-pickle-0.2.3/LICENSE` es una copia sin cambios de [su fuente en el commit 3eb007eae1ca87e3a3683c99a58fd19289abae52](https://github.com/matrix-org/matrix-pickle/blob/3eb007eae1ca87e3a3683c99a58fd19289abae52/LICENSE), el commit declarado en `.cargo_vcs_info.json` de ambos crates. El empaquetador comprueba versión y commit antes de usarlo.

El resto de textos se copia desde las dependencias fijadas al construir cada paquete. `r-efi` incluye su permiso MIT y atribuciones en `AUTHORS`. Los avisos de las bibliotecas estándar proceden de los toolchains instalados. `scripts/collect-licenses.ps1` crea el índice y conserva estos textos dentro de la descarga; un cambio de dependencia que no incluya licencia hace fallar el empaquetado.

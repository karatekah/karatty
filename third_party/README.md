# Supplemental license texts

**English** · [Español](../docs/es/third_party/README.md) · [Italiano](../docs/it/third_party/README.md)

The published crates for `matrix-pickle` and `matrix-pickle-derive` 0.2.3 omit the workspace license file. `matrix-pickle-0.2.3/LICENSE` is an unchanged copy from [source commit 3eb007eae1ca87e3a3683c99a58fd19289abae52](https://github.com/matrix-org/matrix-pickle/blob/3eb007eae1ca87e3a3683c99a58fd19289abae52/LICENSE), the commit declared in both crates' `.cargo_vcs_info.json`. Packaging checks version and commit before using it.

Other license texts are copied from pinned dependencies while building each package. `r-efi` carries its MIT permission and attribution in `AUTHORS`. Standard-library notices come from the installed toolchains. `scripts/collect-licenses.ps1` creates the index and preserves these texts in the download; packaging fails if a dependency has no license text.

Original license and copyright files remain authoritative and are not translated.

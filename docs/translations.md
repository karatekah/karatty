# Documentation languages

**English** · [Español](es/translations.md) · [Italiano](it/translations.md)

English is the primary documentation language and occupies the normal repository paths: `README.md`, root policies and `docs/` guides. Complete Spanish and Italian counterparts live under `docs/es/` and `docs/it/`. Every page links directly to the same page in the other languages.

## Page layout

| English source | Spanish | Italian |
| --- | --- | --- |
| `README.md` | `docs/es/README.md` | `docs/it/README.md` |
| `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md` | Same basename under `docs/es/project/` | Same basename under `docs/it/project/` |
| `docs/install.md` and other guides | `docs/es/install.md` and matching guide | `docs/it/install.md` and matching guide |
| `docs/adr/0003-e2ee-1to1.md` | `docs/es/adr/0003-e2ee-1to1.md` | `docs/it/adr/0003-e2ee-1to1.md` |
| `crypto/README.md`, `third_party/README.md` | Matching subdirectory under `docs/es/` | Matching subdirectory under `docs/it/` |

The page catalog is `docs/languages.json`. It includes release notes, operational plans and inactive legal drafts, not only introductory READMEs. Original license/copyright texts remain authoritative in their original language and are not translated.

## Update a page

1. Update the English source and both complete translations in the same change. Preserve security qualifications, experimental status, numbers, version pins and evidence dates.
2. Keep executable names, flags, environment variables, API routes, JSON fields and signed authentication domains literal. Translate explanations, headings, code comments and obvious example placeholders.
3. Use relative links so a translated guide stays in its chosen language. Links to code/configuration/license files point back to the actual repository file.
4. Keep the three-way language bar after the title, with the current language marked.
5. Run `./scripts/check-docs.ps1` from PowerShell. It checks catalog coverage, three counterparts, language navigation and relative link targets. Then inspect the rendered Markdown/tables/code blocks.

The checker cannot prove semantic equivalence; reviewers must compare the technical meaning. Do not replace a translation with a brief summary or silently fall back to English. Runtime TUI messages are not localized by this documentation change.

## Repository rename and releases

The canonical repository/module is now `karatekah/karatty`. Do not reuse the former repository name, because GitHub redirects old URLs. Existing build paths/environment identifiers keep their documented spelling for compatibility.

Published `v0.1.0-alpha.1` binaries, checksums and tag remain unchanged. Its archived source/docs use the previous internal module and documentation language. Current guides and the Release's language links point to maintained documentation; they do not rewrite that historical source snapshot.

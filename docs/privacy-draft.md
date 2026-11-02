# Karatty privacy policy — review draft

**English** · [Español](es/privacy-draft.md) · [Italiano](it/privacy-draft.md)

**Inactive draft for a future hosted service; do not present it as an active policy.** The current alpha is a local/LAN instance administered by its installer, not a project-hosted service. Each administrator must disclose their instance's processing/copies. `karatty.dev` is selected but unpurchased. Operator identity/contact, actual provider locations, brand checks and legal/operational review are missing. Effective date: `[PENDING]`.

## Proposed controller and contact

The future operator would be `[OPERATOR LEGAL NAME]`, a natural person domiciled in Peru. Privacy/access/rectification/cancellation/objection requests: `[PUBLIC PRIVACY EMAIL]`; support: `[PUBLIC SUPPORT EMAIL]`.

## Data and purposes

- **Account/access:** pseudonym, public SSH key/fingerprint, creation/access dates, session hashes and recovery-code hash. Used for registration, authentication, recovery and abuse prevention. No account email, real name, SSH private key or readable recovery code is collected/stored.
- **Experimental E2EE public keys:** Olm public identity, SSH signature, signed one-time public keys and consumption markers. Distributed only to accepted direct-chat peers and used to prevent one-time-key reuse. The server receives no private identity or ratchet state. Public keys are removed on account deletion or recovery with a replacement SSH key.
- **Encrypted local archive:** readable history, pending sends and ratchet state are encrypted on the device under an independent local passphrase. Opening/using the client requests it and cleans expired entries. Losing it prevents opening the archive; recovery codes cannot replace it. SSH-key replacement migrates readable history and removes old encrypted state after saving the new state. Unconfirmed older sends remain uncertain.
- **Messaging:** E2EE mode stores opaque envelopes, sender/conversation/retry ID/sequence/time plus direct requests/blocks. The older invitation/channel prototype stores readable bodies and is outside the alpha. Metadata support delivery, consent and blocking.
- **Operations:** aggregate request/failure counters and temporary in-memory rate limits. Application operational logs do not contain message bodies, codes or tokens. Infrastructure providers may keep their own service logs.

Pseudonyms do not guarantee absolute anonymity: public keys, contact relationships, timestamps or shared information can identify someone. Participants can read messages delivered to them. Official clients encrypt before sending; the operator retains delivery metadata. Legacy mode exposes readable content and is outside alpha use. HTTPS/WSS protect transport. No project-hosted service is currently operating.

## Providers and transfers

Server/PostgreSQL: `[PROVIDER, REGION, COUNTRY]`. Encrypted backups: `[PROVIDER, REGION, COUNTRY]`. Recovery codes appear once locally rather than being emailed. Complete provider locations/international data flows before hosting.

Proposed statement, subject to verification: personal data are not sold and message content is not used for advertising/model training. `[VERIFY ACTUAL PROVIDER CONTRACTS AND CONFIGURATION]`.

## Retention and deletion

- Messages disappear from history after **720 hours (30 days)**. E2EE-mode physical purging runs at startup/hourly, so removal can lag the access cutoff.
- Direct requests expire after **7 days**, authentication/recovery challenges after **2 minutes**, sessions after **8 hours**. Maintenance removes expired records/revoked sessions.
- Client-driven account deletion removes SSH/E2EE public keys, sessions, memberships and authored messages from active storage, plus local account files under the configured state root on that device. It cannot remotely remove other copies/devices. Other participants' authored messages expire on their own schedule; affected chats become read-only.
- Proposed full backups are encrypted before upload and expire locally/remotely after **3 days**. Deleted/expired data may survive in a backup for at most **7 additional days** under the proposed public policy. Verify actual provider removal before publishing that promise. The local guide does not configure automatic backups.

## Choices and rights

Users can reject direct requests, block accounts and delete their account. Direct access/rectification/cancellation/objection requests to the future privacy contact. A reasonable account-control proof may be requested before replying. See the Peruvian authority's [ARCO-rights guidance](https://www.gob.pe/9270-que-son-los-derechos-arco) and complaint routes.

## Security and changes

SSH keys and the one-time-displayed offline code protect access. Keep them away from others; losing both prevents automatic recovery. Material policy changes would be announced at `[PUBLIC URL]` before taking effect. Code-based recovery replaces SSH/Olm identity and preserves already-readable local history only if the file/passphrase remain available. Contacts must verify the replacement identity.

## Required review before hosting

Review ANPD guidance on [personal-data database registration](https://www.gob.pe/institucion/anpd/pages/8060-inscribir-banco-de-datos-en-el-registro-nacional-de-proteccion-de-datos-personales), including natural-person controllers, and [cross-border data-flow registration](https://www.gob.pe/9253-inscribir-flujo-transfronterizo-de-datos-personales) once provider locations are known. These actions require the controller's personal information; this draft records no completed filing.

The original legal review references [Supreme Decree 016-2024-JUS](https://www.gob.pe/institucion/anpd/normas-legales/6554453-16-2024-jus), recorded as effective March 31, 2025 in ANPD information. The operator must verify applicable requirements before launch. Reading sources or translating this draft is not legal validation or registration.

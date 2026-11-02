# Reporting a vulnerability

**English** · [Español](docs/es/project/SECURITY.md) · [Italiano](docs/it/project/SECURITY.md)

**v0.1.0-alpha.1** is the supported experimental version. It uses local-TUI E2EE for 1:1 direct chats and has [documented tests](docs/verification.md), but no independent audit. The legacy invitation mode and remotely hosted SSH TUI expose readable content to the server and are outside this alpha.

Use **[Security → Report a vulnerability](https://github.com/karatekah/karatty/security/advisories/new)** to send a private report. Do not put exploitable details, keys, tokens or conversations in an issue or discussion. If the private-report button is missing, open an issue without technical details to tell the maintainer; the private channel must be available before sensitive information is requested.

Include the version/commit, platform, reproduction steps and observed impact. Use test data. The maintainer will acknowledge the report and coordinate a fix and disclosure with the reporter. The alpha has no fixed response-time commitment.

There are no stable versions or earlier public versions receiving patches. Apply alpha security updates and read compatibility notes. The project distributes software and does not yet operate a hosted service or public support mailbox.

# Future hosted-service operations

**English** · [Español](es/operations.md) · [Italiano](it/operations.md)

Domain, VPS and providers are deferred, not prerequisites for the [local/LAN alpha](quickstart.md). No infrastructure has been purchased and no remote backups operate. `karatty.dev` was selected but not bought/confirmed with a registrar. Public support/privacy mailboxes do not yet exist. Karatty account registration requires no email.

Working name: **Karatty**. On November 1, 2026, official [Busca tu marca](https://pi.indecopi.gob.pe/buscatumarca/#/inicio) returned **zero precise-name matches** for active marks in classes 9/38/42; broad search returned 200. This is preliminary, not legal clearance. Review similarities, domain and brand before offering branded hosting.

## Infrastructure and budget

The selected future ceiling is USD 20/month. The recorded small [DigitalOcean Basic](https://www.digitalocean.com/pricing/droplets) option has 1 vCPU, 2 GiB RAM and 50 GB SSD; PostgreSQL and Go can share it for controlled alpha capacity. [Cloudflare R2](https://developers.cloudflare.com/r2/pricing/) has a Standard free allowance, and `.dev` is paid annually. Recheck actual prices/quotas before purchase or expansion.

**Recorded November 1, 2026, not a current quote:** DigitalOcean USD 12/month; R2 Standard 10 GB-month, one million Class A/ten million Class B operations monthly free; [Cloudflare Registrar](https://pricing.registrar.cloudflare.com/) `.dev` renewal USD 12.20/year (initial promotions can differ). Baseline estimate: 12 + 12.20/12 ≈ **USD 13.02/month**, before taxes, overages and other services. Registry [RDAP](https://pubapi.registry.google/rdap/domain/karatty.dev) returned 404; that suggested unregistered status then but neither reserves the name nor proves registrar availability. Confirm final purchase price.

[Cloudflare Email Routing](https://developers.cloudflare.com/email-service/platform/pricing/) offers incoming forwarding for possible `support@karatty.dev`/`privacy@karatty.dev`. It [does not send/reply as that domain](https://developers.cloudflare.com/email-service/reference/postmaster/#sending-or-replying-to-an-email-from-your-cloudflare-domain); replies would use the destination mailbox. The operator must select/verify addresses before publishing policies. These addresses are proposals, not active contacts.

## Backups

Use a private dedicated R2 bucket and bucket-scoped rclone credentials. Set `BACKUP_REMOTE=r2:BUCKET_NAME/db-backups`. Generate an `age` identity outside the VPS and keep its private key there; configure only the public `age1…` recipient in `BACKUP_RECIPIENT`.

[backup.sh](../scripts/backup.sh) streams `pg_dump -Fc` through encryption before uploading, retains the local encrypted file only during upload and deletes remote objects older than three days. Schedule daily using systemd with `Persistent=true`.

Also set an R2 [lifecycle rule](https://developers.cloudflare.com/r2/buckets/object-lifecycles/) expiring `db-backups/` after **three days**. Verify application and any versioning/other copies. Cloudflare describes deletion as usually within 24 hours after expiry, potentially later. To promise at most seven additional backup days, monitor actual removal and investigate objects older than four days; lifecycle configuration alone is not an exact physical-deletion guarantee.

Before offering hosting, upload/restore into an empty database whose name begins `termlink_restore_`. On a machine holding the private identity, set `BACKUP_OBJECT`, `AGE_IDENTITY_FILE`, `RESTORE_DATABASE_URL` and run [restore-check.sh](../scripts/restore-check.sh). It refuses other names or existing tables. Verify restored rows/service readability and repeat regularly/after schema changes. Never put the private age identity in the bucket or VPS.

**Local evidence, October 31, 2026:** backup.sh encrypted/uploaded to a local rclone backend; restore-check.sh restored three users/nine messages into a disposable database. The next backup removed a copy marked five days old. This does not test R2 or its lifecycle rule.

## Before opening hosted service

1. Confirm/buy the domain and complete similarity review; establish public support/privacy contact.
2. Review local-TUI 1:1 E2EE, prevent plaintext/channels/remote SSH, verify contacts and test identity loss/rotation.
3. Set up private VPS/bucket, DNS, TLS and administrative access; keep credentials out of Git.
4. Complete/review [privacy](privacy-draft.md)/terms with operator identity/contact, message/backup retention and applicable Peruvian data-protection steps.
5. Test signup, offline-code recovery, acceptance, blocking, deletion, hourly purge, backup, restore and remote removal on purchased infrastructure.
6. Start from a [published verified release](release.md) and test external installation against that hosted service.

Record dated results before announcing hosting. Local-software distribution follows its separate release checklist.

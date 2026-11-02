#!/usr/bin/env bash
set -euo pipefail

# Install as a Certbot deploy hook; ignore certificates for other services.
if [[ " ${RENEWED_DOMAINS:-} " != *' chat.karatty.dev '* ]]; then
  exit 0
fi
: "${RENEWED_LINEAGE:?Certbot must provide RENEWED_LINEAGE}"
test -s "$RENEWED_LINEAGE/fullchain.pem"
test -s "$RENEWED_LINEAGE/privkey.pem"
install -d -o root -g karatty -m 0750 /etc/karatty/tls
install -o root -g karatty -m 0640 "$RENEWED_LINEAGE/fullchain.pem" /etc/karatty/tls/fullchain.pem
install -o root -g karatty -m 0640 "$RENEWED_LINEAGE/privkey.pem" /etc/karatty/tls/privkey.pem
systemctl try-restart karatty.service

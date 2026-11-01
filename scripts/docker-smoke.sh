#!/usr/bin/env bash
set -euo pipefail
export TERMLINK_DB_PASSWORD=local-ci-only-7d3120a9
export TERMLINK_UID="$(id -u)"
export TERMLINK_GID="$(id -g)"
compose=(docker compose -f deployments/compose.yml --profile app)
trap '"${compose[@]}" down' EXIT
go run ./scripts/devcert.go
"${compose[@]}" up -d --build
ready=false
for attempt in {1..30}; do
  if curl --silent --fail --cacert .local/tls/ca.crt https://127.0.0.1:8443/healthz >/dev/null; then ready=true; break; fi
  sleep 1
done
if [[ "$ready" != true ]]; then echo 'Container did not become ready'; exit 1; fi
channel="$("${compose[@]}" exec -T app /server --channel smoke)"
invitation="$("${compose[@]}" exec -T app /server --invite --conversation "$channel" --uses 1)"
echo "::add-mask::$invitation"
mkdir -p .local/smoke bin
ssh-keygen -q -t ed25519 -N '' -f .local/smoke/user
go build -o bin/karatty ./cmd/termlink
printf '%s' "$invitation" | bin/karatty register --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/smoke/user --username smoke --invitation-stdin
printf '%s' 'container persisted message' | bin/karatty send --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/smoke/user --conversation "$channel" --id container-send --message-stdin >.local/smoke/ack.json
bin/karatty history --server https://127.0.0.1:8443 --ca-cert .local/tls/ca.crt --key .local/smoke/user --conversation "$channel" >.local/smoke/history.json
python3 -c 'import json; a=json.load(open(".local/smoke/ack.json")); h=json.load(open(".local/smoke/history.json")); assert h["messages"][0]["id"] == a["id"] and a["seq"] == 1'
echo 'Container TLS + PostgreSQL + native CLI smoke passed.'

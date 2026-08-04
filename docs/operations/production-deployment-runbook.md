---
author: brian
created: "2026-08-04"
document_id: 019f7d55-fb2c-7b8d-9c4a-000000000004
document_type: general
policy_exempt: false
summary: Operator runbook for preparing, deploying, updating, rolling back, and recovering production.
---
# Production Deployment Runbook

Parent index: [Operations](./!INDEX.md)

## Purpose

Provide the smallest complete operator workflow for the implemented production Compose deployment.

## Overview

The deployment host needs Docker Compose, Python 3, GHCR access, the extracted `deploy/production` payload, and a private `.env`. The expected installation path is `/opt/space-rocks`, but the scripts resolve paths relative to their own directory.

## Operating model

1. Obtain the approved release artifact and extract its `production` directory.
2. Copy `.env.multiplayer-production.example` to the deployment directory as `.env`.
3. Fill every required production value, including database and application secrets, internal tokens, Discord configuration where used, allowed origins, and Playit address/ports.
4. Restrict the file: `chmod 600 .env`.
5. Authenticate to `ghcr.io` with `read:packages` if images are private.
6. Run `python3 deploy.py` from `deploy/production`.
7. Run `python3 verify.py` and inspect `docker compose ps`.

`deploy.py` refuses a missing `.env` or values containing `replace-with-`; it runs Compose config validation, image pull, detached startup with orphan removal, and a final service listing.

## Commands/controls

```bash
cd /opt/space-rocks
# authenticate only when GHCR packages are private
docker login ghcr.io
cd deploy/production
python3 deploy.py
python3 verify.py
docker compose --env-file .env -f compose.yaml ps
docker compose --env-file .env -f compose.yaml logs --tail=200
```

For an update, change `SPACE_ROCKS_IMAGE_TAG` and `BUILD_VERSION` together when identifying a release, then rerun deploy and verify. For rollback, select a known-good prior tag, preferably an immutable `sha-<commit>` tag, and repeat the same workflow.

## Runtime state/logs

PostgreSQL and diagnostic data are named-volume state and must not be removed during a routine update. Compose logs are the immediate operator view. The diagnostic volume contains reports and logs under `/data/reports` and `/data/logs`.

## Failure/recovery

- Missing or placeholder environment values: stop; restore the private `.env` from the approved secret-management process.
- GHCR pull failure: verify registry login, package visibility, tag spelling, and host network access.
- API unhealthy: inspect API logs and PostgreSQL health before restarting dependents.
- Game health failure: inspect game-server logs and WebRTC/allowed-origin settings; do not treat a healthy API as game readiness.
- Public traffic failure with healthy local checks: inspect Cloudflare Tunnel routes and Playit assignment/firewall separately.
- Bad release: set the last known-good image tag, rerun deploy, verify, and preserve the failed tag for investigation.

Do not run `docker compose down -v` for recovery unless intentional data destruction has been approved; it removes named volumes.

## Verification

`verify.py` checks Compose state, `http://127.0.0.1:8082/up`, `http://127.0.0.1:8081/health`, and TCP reachability on `127.0.0.1:8083`. These checks do not validate external tunnel routing, Discord OAuth, or public WebRTC candidates; perform those as separate smoke checks when changing those boundaries.

## Related docs

- [Production topology](production-topology.md)
- [Operations controls and verification](operations-controls-and-verification.md)
- [Hosted image release pipeline](hosted-image-release-pipeline.md)
- [Deployment payload README](../../deploy/production/README.md)
- [Production environment example](../../.env.multiplayer-production.example)

## Notes

Do not commit `.env`, paste secret values into tickets, or add credentials to release artifacts. The deployment bundle is server-only and does not require a Git clone.

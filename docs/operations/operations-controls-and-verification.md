---
author: brian
created: "2026-08-04"
document_id: 019f7d55-fb2c-7b8d-9c4a-000000000005
document_type: general
policy_exempt: false
summary: Production controls, runtime inspection, health checks, and operator verification boundaries.
---
# Operations Controls and Verification

Parent index: [Operations](./!INDEX.md)

## Purpose

Define the routine controls and evidence used to decide whether the production deployment is running correctly.

## Overview

Production verification has three layers: Compose/container state, local service health, and external transport reachability. The repository verifier covers the first two and diagnostic listener reachability; operators must separately validate Cloudflare and Playit when those paths change.

## Operating model

Use the deployment directory as the Compose project root. Keep the environment file private and treat named volumes as production state. Prefer targeted restarts or a tag change over destructive teardown.

## Commands/controls

```bash
docker compose --env-file .env -f compose.yaml ps
docker compose --env-file .env -f compose.yaml images
docker compose --env-file .env -f compose.yaml logs --tail=200 api-server game-server diagnostic-aggregator
docker compose --env-file .env -f compose.yaml config --quiet
python3 verify.py
```

The host bindings are `127.0.0.1:8081` game/WebSocket, `127.0.0.1:8082` API, and `127.0.0.1:8083` diagnostics. Game UDP binds the configured local range, default four ports `50000-50003`.

## Runtime state/logs

`postgres-data` is PostgreSQL state. `diagnostic-data` persists diagnostic reports and logs. Container stdout/stderr is inspected through Compose logs; service-specific logging behavior remains in the service owner docs. The Playit agent uses host networking and has no project volume in Compose.

## Failure/recovery

A green Compose state does not prove public routing. For a local failure, inspect the unhealthy service and its dependency, then use a bounded restart or redeploy. For public-only failure, check tunnel route targets, Playit address/port assignment, host firewall, and the advertised WebRTC variables without changing application secrets. Preserve volumes during diagnosis.

## Verification

The implemented verifier:

- runs `docker compose ps`;
- expects API `/up` to return a successful response;
- expects game-server `/health` to return a successful response;
- opens a TCP connection to diagnostic port `8083`;
- prints `Space Rocks hosted services are healthy.` on success.

It does not perform authenticated API checks, database migrations, GHCR provenance checks, Cloudflare checks, Playit checks, or a full multiplayer session. Those are explicit remaining operational checks.

## Related docs

- [Production deployment runbook](production-deployment-runbook.md)
- [Production topology](production-topology.md)
- [API runtime and health](../services/api-server/runtime-and-health.md)
- [Game-server observability](../services/game-server/observability/logging-and-diagnostics.md)
- [Diagnostic aggregator runtime](../services/diagnostic-aggregator/runtime-and-report-flow.md)

## Notes

Health endpoints and ports are facts of the current deployment files. If the Compose payload changes, update this operations folder with the same change.

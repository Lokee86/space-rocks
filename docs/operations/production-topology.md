---
author: brian
created: "2026-08-04"
document_id: 019f7d55-fb2c-7b8d-9c4a-000000000002
document_type: general
policy_exempt: false
summary: Current production service topology, transports, persistence, and security boundaries.
---
# Production Topology

Parent index: [Operations](./!INDEX.md)

## Purpose

Describe the implemented production topology without duplicating service internals.

## Overview

The production bundle is a Docker Compose application named `space-rocks`. It runs PostgreSQL, the Rails API server, the diagnostic aggregator image, the Go game server, and the Playit agent. HTTP containers are loopback-bound on the host and are intended to be reached through existing Cloudflare Tunnel routes. Public realtime UDP is provided by Playit.

## Operating model

```text
Cloudflare Tunnel -> host loopback
  127.0.0.1:8081 -> game-server container :8080 (HTTP and WebSocket)
  127.0.0.1:8082 -> api-server container :3000 (Rails HTTP)
  127.0.0.1:8083 -> diagnostic-aggregator container :8080 (diagnostic HTTP)

Playit agent (host network) -> host UDP range -> game-server UDP range
```

Compose starts PostgreSQL first, then the API after PostgreSQL health, then game-server after API health and diagnostic service start. The diagnostic image is independently deployed in this bundle; it is not the Go process's co-hosted diagnostic implementation.

## Commands/controls

The topology is controlled by `deploy/production/compose.yaml` and `.env`. The image tag is selected by `SPACE_ROCKS_IMAGE_TAG`, defaulting to `p3-hosted`. Do not expose the `.env` contents or place secrets in documentation.

## Runtime state/logs

Named volumes persist PostgreSQL data at `postgres-data` and diagnostic reports/log data at `diagnostic-data`, mounted at `/data` in the diagnostic container. The game-server and API containers have no Compose persistent volume in this bundle; their runtime state is container-local.

## Failure/recovery

A failed dependency blocks dependent startup through Compose health conditions. Restart policy is `unless-stopped` for all runtime services. If the host loses a named volume or the external tunnel/Playit path fails, application container health alone does not prove public reachability; use the verification and recovery procedures in the runbook.

## Verification

Check `docker compose ps` and run `python3 deploy/production/verify.py` from the deployment directory. The verifier checks API `/up`, game-server `/health`, diagnostic TCP reachability, and Compose service state.

## Related docs

- [Production deployment runbook](production-deployment-runbook.md)
- [Operations controls and verification](operations-controls-and-verification.md)
- [API runtime and health](../services/api-server/runtime-and-health.md)
- [Game-server startup](../services/game-server/process/service-startup.md)
- [Diagnostic aggregator runtime](../services/diagnostic-aggregator/runtime-and-report-flow.md)

## Notes

The four local UDP ports are persistent Pion ICE UDP multiplexers, not one-port-per-player allocations. The advertised public address and four-port range must match the Playit assignment. `WEBRTC_ICE_SERVERS` is optional for this Playit path.

---
author: brian
created: "2026-08-04"
document_id: 019f7d55-fb2c-7b8d-9c4a-000000000003
document_type: general
policy_exempt: false
summary: Current hosted image publishing workflow and server-only deployment bundle.
---
# Hosted Image Release Pipeline

Parent index: [Operations](./!INDEX.md)

## Purpose

Document how production images and the deployment payload are built, validated, published, and consumed.

## Overview

`.github/workflows/publish-hosted-images.yml` runs on manual dispatch or pushes to `deploy/p3-hosted-stack` that touch the workflow, production deployment files, participating services, or shared files. It first validates Compose with temporary validation values, then publishes three private GHCR images and packages the `deploy/production` directory.

## Operating model

The publish job builds and pushes:

- `ghcr.io/lokee86/space-rocks-game-server:p3-hosted`
- `ghcr.io/lokee86/space-rocks-api-server:p3-hosted`
- `ghcr.io/lokee86/space-rocks-diagnostic-aggregator:p3-hosted`

Each image also receives `sha-${{ github.sha }}`. Build contexts and Dockerfiles are declared in the workflow. OCI source and revision labels are attached, and GitHub Actions cache scopes are per image.

The deployment-bundle job creates `space-rocks-p3-hosted-deploy.tar.gz` from `deploy/production` and uploads it as the `space-rocks-p3-hosted-deploy` artifact. The bundle contains deployment files, not the repository source or secrets.

## Commands/controls

The workflow requires `contents: read` and `packages: write`. The server must authenticate to GHCR with `read:packages` if packages remain private. On the server, choose an image tag in `.env`; the Compose file applies that tag to all three Space Rocks images.

## Runtime state/logs

GitHub Actions retains the uploaded bundle as a workflow artifact. GHCR retains the tagged image manifests. The deployment host owns the extracted Compose files and its private `.env`; neither the artifact nor repository stores production secret values.

## Failure/recovery

If Compose validation fails, no image or bundle job proceeds. If one matrix image fails, the matrix is fail-fast false, so inspect all matrix results before deploying. A failed or incorrect release can be isolated by selecting a prior immutable `sha-<commit>` tag in `.env`, provided that tag is available in GHCR.

## Verification

Before deployment, verify the workflow's validation job passed, all three image tags exist, and the bundle artifact is present. After deployment, run `deploy.py` and `verify.py`; image identity can be checked with `docker compose images` and the selected `SPACE_ROCKS_IMAGE_TAG`.

## Related docs

- [Production deployment runbook](production-deployment-runbook.md)
- [Production topology](production-topology.md)
- [Production deployment payload](../../deploy/production/README.md)
- [Game-server Dockerfile](../../services/game-server/Dockerfile)
- [API-server Dockerfile](../../services/api-server/Dockerfile)
- [Diagnostic Dockerfile](../../services/diagnostic-aggregator/Dockerfile)

## Notes

The workflow's validation `.env` is intentionally disposable and uses validation-only values. It is not a production configuration template.

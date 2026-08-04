---
author: brian
created: "2026-08-04"
document_id: 019f7d55-fb2c-7b8d-9c4a-000000000001
document_type: general
policy_exempt: false
summary: Canonical production deployment and hosted release operations documentation.
---
# Operations

Parent index: [Documentation](../!INDEX.md)

## Purpose

This folder owns the current production deployment, hosted image release, and operator workflow for Space Rocks.

## Ownership

Operations docs describe deployment topology, release artifacts, environment ownership, runtime controls, verification, rollback, and recovery. Service docs remain authoritative for service implementation behavior, health endpoint semantics, and diagnostic internals.

## Direct Files
<!-- doc-ledger:files:start -->

- [Production topology](production-topology.md) - Production services, transports, ports, volumes, and security boundaries.
- [Hosted image release pipeline](hosted-image-release-pipeline.md) - GitHub Actions image publishing and deployment-bundle ownership.
- [Production deployment runbook](production-deployment-runbook.md) - Environment preparation, deploy, verify, update, rollback, and recovery workflow.
- [Operations controls and verification](operations-controls-and-verification.md) - Runtime state, logs, controls, failure handling, and verification expectations.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Documentation](../!INDEX.md)
- [Production deployment payload](../../deploy/production/README.md)
- [Developer guide](../developer.md)

## Notes

These documents describe the implemented `deploy/production` payload and `publish-hosted-images.yml`; they do not define future hosting or service architecture.

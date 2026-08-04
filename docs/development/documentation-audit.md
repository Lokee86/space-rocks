---
author: brian
created: "2026-08-04"
document_id: 6afa6411-b884-4af5-9cf7-4a18248fffb2
document_type: general
policy_exempt: false
summary: Records the August 2026 Space Rocks documentation audit, remediation scope, verification evidence, and remaining debt.
---
# Documentation Audit

Parent index: [Development](./!INDEX.md)

## Purpose

This document records the repository-wide documentation audit initiated on August 4, 2026 and the evidence used to remediate it.

## Overview

The audit compares current implementation and operational surfaces against Space Rocks documentation policy and the canonical `engineering-standards` game profile. It distinguishes structural navigation from semantic currency: a complete index is useful, but it does not prove that current behavior is accurately documented.

## Initial findings

The initial inspection found:

- all documentation folders had `!INDEX.md` files and direct-file/direct-folder index coverage;
- no duplicate `document_id` values were detected;
- the repository lacked `docs-standard.json`, a generated standards snapshot, a maintainer map, an implementation coverage map, and a behavioral-contract matrix;
- the diagnostic-aggregator canonical service documentation described a hosted-only runtime even though a standalone executable, Docker image, and production Compose service existed;
- `deploy/production/` and image-publishing workflows had no canonical operations owner under `docs/`;
- implemented team packages were represented primarily by planning material rather than focused current service owners;
- broken relative links and legacy document-shape debt existed across the corpus;
- root agent guidance still described the Rails API server as a planned placeholder.

## Remediation scope

This audit owns:

1. shared engineering-standards snapshot and checker adoption;
2. repository profile, coverage, change-impact, and Pitlord documentation-policy integration;
3. maintainer-map and behavioral-contract ownership;
4. correction of dangerous factual drift;
5. canonical operations documentation for production deployment and release publishing;
6. graduation of implemented behavior from planning where current owners were missing;
7. broken-link and required-section remediation or explicit retained baseline debt;
8. CI enforcement for structural and change-impact checks.

## Audit method

The review uses:

- repository tree and package/executable inventory;
- current service, protocol, data, operations, planning, limits, and agent documents;
- executable entry points, Dockerfiles, Compose configuration, workflows, source-of-truth schemas, and tests;
- the shared documentation checker from `.standards/docs_policy/`;
- repository Pitlord policy;
- focused semantic inspection of high-risk claims such as `current`, `not implemented`, `deferred`, `temporary`, `hosted`, and `standalone`.

Generated, vendored, addon, scene, and large test files are legitimate size and coverage exceptions where they are already owned by a focused subsystem. Documentation is exempt from production source-file line limits.

## Status

The audit remains active until the final branch verification records:

- structural checker result;
- changed-from change-impact result;
- Pitlord result;
- repository test result appropriate to documentation/tooling changes;
- corrected factual owners and any explicitly retained documentation debt.

## Related docs

- [Documentation standard snapshot](../../.standards/docs/documentation-standard.md)
- [Documentation policy](../documentation-policy.md)
- [Documentation procedure](../documentation-procedure.md)
- [Maintainer map](../maintainer-map.md)
- [Documentation coverage](documentation-coverage.md)
- [Behavioral-contract matrix](behavioral-contract-matrix.md)

## Notes

Do not change this document to claim the corpus is complete merely because structural checks pass. Semantic currency requires inspection against current implementation and operations.

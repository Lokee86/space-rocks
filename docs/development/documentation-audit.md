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

The repository-wide remediation is complete on `docs/full-audit`.

Completed work includes:

- synchronized the canonical engineering-standards snapshot into `.standards/`;
- added `docs-standard.json`, a maintainer map, implementation coverage, a behavioral-contract matrix, and this audit owner;
- wired the shared structural checker and pull-request change-impact checker into CI;
- corrected the diagnostic-aggregator from a false hosted-only model to its actual standalone-plus-hosted-adapter architecture;
- added canonical production topology, image publishing, deployment, verification, update, rollback, and recovery documentation;
- graduated implemented server/client team ownership out of planning while leaving unmerged P4 mechanics in planning;
- corrected broken relative links, stale agent guidance, universal section debt, and the Bruno collection's nonexistent request inventory;
- corrected the stale player-data contract-test limitation;
- retained no documentation baseline suppressions.

Verification evidence:

- shared full documentation checker: passed with zero findings, zero baselined findings, and zero stale baseline entries;
- shared changed-from documentation-impact checker against `main`: passed with zero findings;
- Pitlord `v0.1.1`: passed 25 repository policy rules with no violations;
- Python repository/tooling suite: 343 passed;
- data-sync source validation: passed;
- generated constants, packet, realtime-wire, and drop-table checks: passed.

The generated shared Pitlord documentation policy currently contains a `require_content` rule that is unsupported by Pitlord `v0.1.0`, `v0.1.1`, `v0.1.2`, and the reachable remote `main`. Space Rocks therefore mirrors the supported required-path rules in its local policy, keeps the required `AGENTS.md` language in source, and uses the shared Python checker as the authoritative documentation gate. The exact content assertion can move back into the included shared policy when a compatible Pitlord release exists.

## Related docs

- [Documentation standard snapshot](../../.standards/docs/documentation-standard.md)
- [Documentation policy](../documentation-policy.md)
- [Documentation procedure](../documentation-procedure.md)
- [Maintainer map](../maintainer-map.md)
- [Documentation coverage](documentation-coverage.md)
- [Behavioral-contract matrix](behavioral-contract-matrix.md)

## Notes

Do not change this document to claim the corpus is complete merely because structural checks pass. Semantic currency requires inspection against current implementation and operations.

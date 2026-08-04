# Space Rocks architecture policy

`policy.json` is the executable repository architecture policy evaluated by Pitlord.
It replaces the former Python architecture-guard engine and the ad hoc player-data and
server-devtools boundary scans.

Run from the repository root:

```bash
pitlord check --repo . --policy tools/pitlord/policy.json
```

The current policy uses repository-only content and path rules, so this gate does not
require a Lexicon or Arcana snapshot. Semantic dependency rules may be added to separate
included policy files when the CI graph-refresh workflow is enabled.

The generated shared documentation policy is retained at
`.standards/policies/documentation-core.json`. Its required-path rules are mirrored in the
repository policy. The current published Pitlord releases do not yet support the shared
policy's `require_content` rule, so the exact `AGENTS.md` documentation-language assertion
is present in source and reviewed by the documentation audit but is not included in the
Pitlord composition root until a compatible release is available. The shared Python
documentation checker remains the authoritative index, link, section, coverage, and
change-impact gate.

Keep rules narrow, evidence-based, and tied to an existing documented invariant. Contract,
generation, documentation, packaging, and behavioral tests remain in their owning test
systems rather than being folded into Pitlord.

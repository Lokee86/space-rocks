# Space Rocks Pitlord Policy

Space Rocks runs Pitlord as a full repository and semantic architecture gate.

## Policy layout

- `policy.json` is the composed entry point.
- `repository.json` contains deterministic content and required/forbidden path rules.
- `semantic.json` declares ownership areas and Arcana-backed dependency, ownership, and cycle rules.
- `run.sh` refreshes Lexicon, synchronizes Arcana, validates the composed policy, and runs Pitlord.

The semantic policy covers the client script domains, game-server internal packages, player-data, diagnostic-aggregator, API server, and shared Go contracts. Go areas include both physical source paths and Lexicon's normalized `@internal/...` package identities so imports are evaluated against their real graph targets.

## Run locally

Pitlord, Lexicon, and Arcana must be available on `PATH`, or supplied through `PITLORD`, `LEXICON`, and `ARCANA` environment variables:

```bash
bash tools/pitlord/run.sh
```

When Lexicon is run from a source checkout, point it at the adapters directory:

```bash
LEXICON_ADAPTERS=/path/to/grimoire/lexicon/adapters \
  bash tools/pitlord/run.sh
```

The first run initializes `.lexicon/` and builds a complete graph. Later runs use Lexicon's incremental scan and Arcana's incremental synchronization. Both state directories are ignored by Git.

Optional settings:

- `LEXICON_LANGUAGES` limits first-time initialization to a comma-separated language set.
- `LEXICON_MAX_WORKERS` bounds Lexicon concurrency.
- `PITLORD_TIMEOUT` changes the graph evaluation timeout from its five-minute default.
- `PITLORD_POLICY` selects another composed policy entry point.

## CI

The repository-check job installs Pitlord `v0.1.1`, builds Lexicon and Arcana from the pinned Grimoire commit recorded in `.github/workflows/ci.yml`, installs the TypeScript adapter dependencies, and invokes `tools/pitlord/run.sh` through the shared repository-check runner.

The pinned Grimoire commit makes semantic results reproducible. Updating that pin requires a clean full scan, review of any changed graph evidence, and successful repository checks.

## Rule changes

Keep rules narrow and tied to an owned invariant. Do not suppress a real dependency by weakening an area selector or excluding broad source trees. Correct the ownership direction when the graph exposes a genuine violation. Baselines are reserved for explicitly accepted legacy evidence, not ordinary policy calibration.

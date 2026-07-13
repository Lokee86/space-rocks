# Architecture guard

The architecture guard is a generic, configuration-driven static check. Run it from the repository root with:

```bash
python tools/architecture_guard/main.py
```

Rules live in `tools/architecture_guard/rules.toml`. The engine supports:

- `content_rules` for literal or regular-expression matches
- `include` and `exclude` repository-relative globs
- `path_rules` for required or forbidden paths

Each rule has an `id`, a pattern, and an optional message. Findings include the rule ID, relative file, line where applicable, and message. Repository traversal skips generated or nested checkout areas such as `.git`, `.worktrees`, `.workingtrees`, `.godot`, `node_modules`, and `__pycache__`.

Add a narrow invariant by adding one rule to `rules.toml`, selecting the smallest include glob and any legitimate exclude globs. Keep ownership exceptions explicit in configuration rather than adding repository-specific logic to the Python engine. Add a focused test in `tools/tests/test_architecture_guard.py` when the rule needs matching or exclusion coverage.

The repository architecture-rule entry point is also covered by `tools/tests/test_architecture_rules.py` and is invoked by `tools/ci/run_repo_checks.sh`.

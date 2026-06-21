# Repo Hygiene
Parent index: [Agent](./!INDEX.md)

## Purpose

This doc owns agent-facing repo safety rules.

## Overview

Space Rocks worktrees may include unrelated user/editor changes, so agents must handle repository state carefully while keeping scoped work focused.

## Rules

- Assume the repo may be dirty.
- Do not clean or revert unrelated user/editor changes casually.
- Godot scene/import diffs may be noisy.
- Godot may rewrite `uid`, `unique_id`, offsets, imports, and scene metadata.
- Generated recordings and build artifacts should not be committed.
- Avoid committing `*.avi`, `tmp/`, `*/tmp/`, and `client/.godot/`.
- The older `space-rocks-(4.3)/` copy is not the active project.
- Avoid broad cleanup during scoped work.

## Related docs

- [Documentation Editing](./documentation-editing.md)
- [Godot Editing](./godot-editing.md)

## Notes

This doc owns permanent repo hygiene; temporary status belongs in `current-context.md`.

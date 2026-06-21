# Prompting And Reporting
Parent index: [Agent](./!INDEX.md)

## Purpose

This doc owns prompt/report expectations for implementation agents.

## Overview

Space Rocks agent work should stay small, bounded, and easy to review.

## Rules

- Prompts should be small and specific.
- Implementation tasks should usually target under two minutes of agent work.
- Each prompt should have one clear edit goal.
- Avoid mixing unrelated refactors.
- Stop if the task balloons.
- Reports should include changed files.
- Reports should mention unexpected files touched.
- Numbered completion headings should be placed at the bottom when requested.
- Command output should only be reported when the command was actually run.

## Related docs

- [Testing](./testing.md)
- [Documentation Editing](./documentation-editing.md)
- [Repo Hygiene](./repo-hygiene.md)

## Notes

This doc does not replace task-specific prompt instructions.

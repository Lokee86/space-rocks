# Generated Files
Parent index: [Agent](./!INDEX.md)

## Purpose

This doc tells agents how to handle generated files safely.

## Overview

Generated outputs must be changed through their source-of-truth files or generators, not by editing the generated artifacts directly.

## Rules

- Do not hand-edit generated packet files.
- Do not hand-edit generated constants files.
- Do not hand-edit generated collision JSON as a convenience.
- Find the source-of-truth file first.
- Use data docs for pipeline details.
- Packet and constants changes flow through shared source files and data-sync.
- Collision-shape data comes from the Godot exporter.

## Related docs

- [Data](../data/!INDEX.md)
- [Testing](./testing.md)
- [Documentation Editing](./documentation-editing.md)

## Notes

This doc is an agent safety guide, not the canonical data pipeline reference.

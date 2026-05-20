#!/usr/bin/env python3

from __future__ import annotations

import argparse
from pathlib import Path

from constants_conversion import (
    default_output_path,
    load_schema,
    parse_gdscript_constants,
    render_go_from_gdscript,
    repo_root,
)


def main() -> None:
    root = repo_root()
    parser = argparse.ArgumentParser(description="Generate a .go constants file from a .gd constants file.")
    parser.add_argument(
        "input",
        type=Path,
        nargs="?",
        help="GDScript constants file to read. Defaults to client/scripts/constants.gd.",
    )
    parser.add_argument(
        "--output",
        type=Path,
        help="Output path. Defaults to the input file name with a .go extension.",
    )
    parser.add_argument(
        "--schema",
        type=Path,
        default=root / "shared/constants/constants.json",
        help="Constants schema used for names and grouping.",
    )
    args = parser.parse_args()

    input_path = args.input or root / "client/scripts/constants.gd"
    output_path = args.output
    if output_path is None:
        if args.input is None:
            output_path = root / "services/game-server/internal/constants/constants.go"
        else:
            output_path = default_output_path(input_path, ".go")

    schema = load_schema(args.schema)
    constants = parse_gdscript_constants(input_path)
    lines = render_go_from_gdscript(schema, constants)

    output_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"generated {output_path}")


if __name__ == "__main__":
    main()

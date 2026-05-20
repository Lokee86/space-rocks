#!/usr/bin/env python3

from __future__ import annotations

import argparse
from pathlib import Path

from constants_conversion import (
    default_output_path,
    load_schema,
    parse_go_constants,
    render_gdscript_from_go,
    repo_root,
)


def main() -> None:
    root = repo_root()
    parser = argparse.ArgumentParser(description="Generate a .gd constants file from a .go constants file.")
    parser.add_argument(
        "input",
        type=Path,
        nargs="?",
        help="Go constants file to read. Defaults to services/game-server/internal/constants/constants.go.",
    )
    parser.add_argument(
        "--output",
        type=Path,
        help="Output path. Defaults to the input file name with a .gd extension.",
    )
    parser.add_argument(
        "--schema",
        type=Path,
        default=root / "shared/constants/constants.json",
        help="Constants schema used for names and grouping.",
    )
    args = parser.parse_args()

    input_path = args.input or root / "services/game-server/internal/constants/constants.go"
    output_path = args.output
    if output_path is None:
        if args.input is None:
            output_path = root / "client/scripts/constants.gd"
        else:
            output_path = default_output_path(input_path, ".gd")

    schema = load_schema(args.schema)
    constants = parse_go_constants(input_path)
    lines = render_gdscript_from_go(schema, constants)

    output_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"generated {output_path}")


if __name__ == "__main__":
    main()

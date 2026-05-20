from __future__ import annotations

import pytest

from data_sync.block_io import BlockIOError, extract_block, find_all_blocks, find_block, replace_block


def test_extracts_block() -> None:
    text = """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 420.0
const BulletSpeed = 900.0
// data-sync:end constants.gameplay
""".lstrip()

    assert extract_block(text, "constants.gameplay") == (
        "const PlayerSpeed = 420.0\n"
        "const BulletSpeed = 900.0\n"
    )


def test_replaces_block() -> None:
    text = """
before
// data-sync:start constants.gameplay
old
// data-sync:end constants.gameplay
after
""".lstrip()

    updated = replace_block(text, "constants.gameplay", "new_a\nnew_b")

    assert updated == """
before
// data-sync:start constants.gameplay
new_a
new_b
// data-sync:end constants.gameplay
after
""".lstrip()


def test_supports_gdscript_markers() -> None:
    text = """
# data-sync:start packets
const PACKET_INPUT := 100
# data-sync:end packets
""".lstrip()

    assert extract_block(text, "packets") == "const PACKET_INPUT := 100\n"


def test_missing_start_marker_errors() -> None:
    text = """
content
// data-sync:end constants.gameplay
""".lstrip()

    with pytest.raises(BlockIOError, match="missing data-sync start marker"):
        find_block(text, "constants.gameplay")


def test_missing_end_marker_errors() -> None:
    text = """
// data-sync:start constants.gameplay
content
""".lstrip()

    with pytest.raises(BlockIOError, match="missing data-sync end marker"):
        find_block(text, "constants.gameplay")


def test_missing_block_errors() -> None:
    with pytest.raises(BlockIOError, match="missing data-sync block"):
        find_block("no markers here\n", "constants.gameplay")


def test_duplicate_marker_errors() -> None:
    text = """
// data-sync:start constants.gameplay
one
// data-sync:end constants.gameplay
// data-sync:start constants.gameplay
two
// data-sync:end constants.gameplay
""".lstrip()

    with pytest.raises(BlockIOError, match="duplicate data-sync block"):
        find_all_blocks(text)


def test_preserves_surrounding_file_content() -> None:
    prefix = "package constants\n\nfunc untouched() {}\n"
    suffix = "\nfunc alsoUntouched() {}\n"
    text = (
        prefix
        + "// data-sync:start constants.gameplay\n"
        + "old\n"
        + "// data-sync:end constants.gameplay\n"
        + suffix
    )

    updated = replace_block(text, "constants.gameplay", "new")

    assert updated.startswith(prefix)
    assert updated.endswith(suffix)


def test_preserves_markers() -> None:
    text = """
# data-sync:start constants.client
old
# data-sync:end constants.client
""".lstrip()

    updated = replace_block(text, "constants.client", "new")

    assert "# data-sync:start constants.client\n" in updated
    assert "# data-sync:end constants.client\n" in updated


def test_refuses_to_rewrite_outside_managed_blocks() -> None:
    text = """
before
// data-sync:start constants.gameplay
old
// data-sync:end constants.gameplay
after
""".lstrip()

    updated = replace_block(text, "constants.gameplay", "before\nnew\nafter")

    assert updated == """
before
// data-sync:start constants.gameplay
before
new
after
// data-sync:end constants.gameplay
after
""".lstrip()

from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.cli import parse_args


def test_push_requires_exactly_one_operation() -> None:
    args = parse_args(["-push", "-constants", "-go"])

    assert args.operation == "push"
    assert args.domains == ("constants",)
    assert args.languages == ("go",)


def test_push_accepts_drop_tables_domain() -> None:
    args = parse_args(["-push", "-drop-tables", "-go"])

    assert args.operation == "push"
    assert args.domains == ("drop_tables",)
    assert args.languages == ("go",)


def test_multiple_operations_are_invalid() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-push", "-pull", "-constants", "-go"])

    assert exc.value.code == 2


def test_no_operation_is_invalid() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-constants", "-go"])

    assert exc.value.code == 2


@pytest.mark.parametrize("operation", ["-push", "-pull", "-diff", "-check"])
def test_sync_operations_require_a_domain(operation: str) -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args([operation, "-go"])

    assert exc.value.code == 2


@pytest.mark.parametrize("operation", ["-push", "-pull", "-diff", "-check"])
def test_sync_operations_require_a_language(operation: str) -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args([operation, "-constants"])

    assert exc.value.code == 2


def test_drop_tables_requires_go() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-push", "-drop-tables"])

    assert exc.value.code == 2


def test_drop_tables_rejects_gds_language() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-push", "-drop-tables", "-gds"])

    assert exc.value.code == 2


def test_pull_allows_one_language() -> None:
    args = parse_args(["-pull", "-constants", "-ts"])

    assert args.operation == "pull"
    assert args.languages == ("ts",)


def test_pull_rejects_multiple_languages() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-pull", "-constants", "-go", "-gds"])

    assert exc.value.code == 2


def test_validate_can_run_alone() -> None:
    args = parse_args(["-validate"])

    assert args.operation == "validate"
    assert args.domains == ()
    assert args.languages == ()


def test_validate_allows_optional_filters() -> None:
    args = parse_args(["-validate", "-constants", "-packets", "-go"])

    assert args.operation == "validate"
    assert args.domains == ("constants", "packets")
    assert args.languages == ("go",)


def test_validate_allows_player_data_domain() -> None:
    args = parse_args(["-validate", "-player_data"])

    assert args.operation == "validate"
    assert args.domains == ("player_data",)
    assert args.languages == ()


def test_player_data_is_validate_only() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-push", "-player_data", "-go"])

    assert exc.value.code == 2


def test_config_and_sot_options() -> None:
    args = parse_args(
        [
            "-diff",
            "-constants",
            "-go",
            "-config",
            "tools/data_sync/config.toml",
            "-sot",
            "shared/data_sync.toml",
        ]
    )

    assert args.config == Path("tools/data_sync/config.toml")
    assert args.sot == Path("shared/data_sync.toml")


def test_multiple_domains_and_languages_are_preserved_in_order() -> None:
    args = parse_args(["-check", "-packets", "-constants", "-ts", "-go", "-gds"])

    assert args.domains == ("constants", "packets")
    assert args.languages == ("go", "gds", "ts")


def test_realtime_wire_accepts_json_and_docs_outputs() -> None:
    args = parse_args(["-diff", "-realtime-wire", "-json", "-docs"])

    assert args.languages == ()
    assert args.output_kinds == ("json", "docs")


def test_observability_accepts_all_output_kinds() -> None:
    args = parse_args(["-diff", "-observability", "-go", "-gds", "-ruby", "-json", "-docs"])

    assert args.domains == ("observability",)
    assert args.languages == ("go", "gds", "ruby")
    assert args.output_kinds == ("json", "docs")


def test_observability_ruby_is_a_language() -> None:
    args = parse_args(["-push", "-observability", "-ruby"])

    assert args.languages == ("ruby",)


def test_json_output_requires_realtime_wire() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-diff", "-packets", "-json"])

    assert exc.value.code == 2


def test_ruby_output_requires_observability() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-diff", "-realtime-wire", "-ruby"])

    assert exc.value.code == 2


def test_realtime_wire_rejects_ts() -> None:
    with pytest.raises(SystemExit) as exc:
        parse_args(["-diff", "-realtime-wire", "-ts"])

    assert exc.value.code == 2

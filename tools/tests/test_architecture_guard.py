from pathlib import Path

from tools.architecture_guard import run_guard


def _config(tmp_path: Path, text: str) -> Path:
    config = tmp_path / "rules.toml"
    config.write_text(text, encoding="utf-8")
    return config


def test_literal_matching_reports_file_and_line(tmp_path: Path) -> None:
    (tmp_path / "src.py").write_text("ok\nBAD literal\n", encoding="utf-8")
    config = _config(tmp_path, '[[content_rules]]\nid = "literal"\npattern = "BAD literal"\n')

    findings = run_guard(tmp_path, config)

    assert [(item.rule_id, item.path, item.line) for item in findings] == [("literal", "src.py", 2)]


def test_regex_matching(tmp_path: Path) -> None:
    (tmp_path / "src.py").write_text("value = secret_123\n", encoding="utf-8")
    config = _config(tmp_path, '[[content_rules]]\nid = "regex"\nkind = "regex"\npattern = "secret_[0-9]+"\n')

    assert len(run_guard(tmp_path, config)) == 1


def test_include_and_exclude_globs(tmp_path: Path) -> None:
    (tmp_path / "src.py").write_text("BAD\n", encoding="utf-8")
    vendor = tmp_path / "vendor.py"
    vendor.write_text("BAD\n", encoding="utf-8")
    config = _config(tmp_path, '[[content_rules]]\nid = "scope"\npattern = "BAD"\ninclude = ["*.py"]\nexclude = ["vendor.py"]\n')

    findings = run_guard(tmp_path, config)

    assert [(item.path, item.line) for item in findings] == [("src.py", 1)]


def test_required_path(tmp_path: Path) -> None:
    config = _config(tmp_path, '[[path_rules]]\nid = "required"\nkind = "required"\npattern = ".github/workflows/ci.yml"\n')

    findings = run_guard(tmp_path, config)

    assert findings[0].rule_id == "required"
    assert "missing" in findings[0].message


def test_forbidden_path(tmp_path: Path) -> None:
    (tmp_path / "forbidden.txt").write_text("", encoding="utf-8")
    config = _config(tmp_path, '[[path_rules]]\nid = "forbidden"\nkind = "forbidden"\npattern = "forbidden.txt"\n')

    assert run_guard(tmp_path, config)[0].path is None


def test_clean_success(tmp_path: Path) -> None:
    (tmp_path / "clean.py").write_text("print('clean')\n", encoding="utf-8")
    config = _config(tmp_path, '[[content_rules]]\nid = "clean"\npattern = "BAD"\n')

    assert run_guard(tmp_path, config) == []


def test_nested_worktree_content_is_not_scanned(tmp_path: Path) -> None:
    nested = tmp_path / ".workingtrees" / "nested"
    nested.mkdir(parents=True)
    (nested / "bad.py").write_text("BAD\n", encoding="utf-8")
    config = _config(tmp_path, '[[content_rules]]\nid = "excluded"\npattern = "BAD"\n')

    assert run_guard(tmp_path, config) == []

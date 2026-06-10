from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


def read_text(relative_path: str) -> str:
    return (ROOT / relative_path).read_text(encoding="utf-8")


def test_info_server_does_not_import_write_helpers():
    text = read_text("tools/space-rocks-mcp/server-info-next.js")

    assert "repo_write_tools" not in text
    assert "engineforge_write_tools" not in text
    assert "allowed_commands" not in text


def test_readonly_tools_remain_readonly():
    text = read_text("tools/space-rocks-mcp/shared/repo_readonly_tools.js")

    assert "writeFile" not in text
    assert "spawn" not in text
    assert "exec" not in text
    assert "runAllowedCommand" not in text


def test_write_server_imports_write_helpers():
    text = read_text("tools/space-rocks-mcp/server-write.js")

    assert "repo_write_tools" in text
    assert "engineforge_write_tools" in text

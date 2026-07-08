from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

def read_text(relative_path: str) -> str:
    return (ROOT / relative_path).read_text(encoding="utf-8")

def test_info_server_blocks_write_tools_but_allows_hermes_cli() -> None:
    text = read_text("tools/space-rocks-mcp/server-info-next.js")

    # Write tools are blocked
    assert "repo_write_tools" not in text
    assert "engineforge_write_tools" not in text
    assert "allowed_commands" not in text

    # Hermes CLI tools are explicitly allowed
    assert "hermes_tools" in text


def test_readonly_tools_remain_readonly() -> None:
    text = read_text("tools/space-rocks-mcp/shared/repo_readonly_tools.js")

    assert "writeFile" not in text
    assert "spawn" not in text
    assert "exec" not in text
    assert "runAllowedCommand" not in text


def test_hermes_tools_is_session_oriented_not_arbitrary_runner() -> None:
    """Hermes MCP access is session-oriented, not arbitrary command execution.

    The Hermes module provides bounded session tools:
    - hermes_ping, hermes_help for diagnostics
    - hermes_session_send, hermes_session_status, hermes_sessions_list for session management

    No generic hermes_run or arbitrary CLI runner exists.
    """
    text = read_text("tools/space-rocks-mcp/shared/hermes_tools.js")

    # spawn is used for child process execution
    assert "spawn" in text

    # exec is NOT used (must use spawn, not exec)
    assert "exec" not in text

    # shell: false is required for Hermes CLI execution
    assert "shell: false" in text

    # hermes command is used
    assert '"hermes"' in text

    # No general shell commands allowed
    assert "bash" not in text
    assert "powershell" not in text
    assert "cmd" not in text

    # No allowed_commands import
    assert "allowed_commands" not in text

    # Must use server.registerTool, not server.addTool
    assert "server.addTool" not in text
    assert "server.registerTool" in text

    # hermes_run is NOT present (removed, no generic runner)
    assert "hermes_run" not in text

    # Session-oriented tools are present
    assert "hermes_session_send" in text
    assert "hermes_session_status" in text
    assert "hermes_sessions_list" in text

    # Session send uses --continue and -z for bounded session interaction
    assert "--continue" in text
    assert "sessions" in text
    assert "status" in text
    assert '"-z"' in text


def test_write_server_imports_write_helpers() -> None:
    text = read_text("tools/space-rocks-mcp/server-write.js")

    assert "repo_write_tools" in text
    assert "engineforge_write_tools" in text
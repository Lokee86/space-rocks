from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[2]

def read_text(relative_path: str) -> str:
    return (ROOT / relative_path).read_text(encoding="utf-8")

def test_info_server_blocks_write_tools_but_allows_hermes_cli() -> None:
    text = read_text("tools/space-rocks-mcp/server-info-next.js")

    # Write tools are blocked
    assert "repo_write_tools" not in text
    assert "engineforge_write_tools" not in text
    assert "allowed_commands" not in text
    assert "workspace_writes" not in text
    assert "apply_repo_file_edits" not in text
    assert "write_repo_file" not in text
    assert "replace_in_repo_file" not in text
    assert "restricted_command_tools" not in text
    assert "command_job_start" not in text
    assert "list_workspace_commands" not in text

    # Hermes CLI tools are explicitly allowed
    assert "hermes_tools" in text

    # Process-job lifecycle tools are registered on the Info server
    assert "job_tools" in text
    assert "registerJobTools" in text


def test_readonly_tools_remain_readonly() -> None:
    text = read_text("tools/space-rocks-mcp/shared/repo_readonly_tools.js")

    assert "writeFile" not in text
    assert "spawn" not in text
    assert "exec" not in text
    assert "runAllowedCommand" not in text


def test_hermes_tools_allow_hermes_cli_args_but_not_general_shells() -> None:
    """Hermes MCP access may pass Hermes CLI args, but not arbitrary shell commands.

    The Hermes module provides a bounded Hermes CLI surface:
    - hermes_run for arbitrary Hermes CLI args and optional stdin
    - hermes_ping, hermes_help for diagnostics
    - hermes_session_send, hermes_session_status, hermes_sessions_list for session management

    The boundary still forbids shell execution wrappers and repo write imports.
    """
    text = read_text("tools/space-rocks-mcp/shared/hermes_tools.js")

    # spawn is used for child process execution
    assert "spawn" in text

    # General exec wrappers are forbidden, but the bounded executable lookup is allowed.
    assert "execFileSync" in text
    assert re.search(r"\bexec\s*\(", text) is None

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

    # hermes_run is present as a Hermes CLI runner, not a general shell
    assert "hermes_run" in text

    # Session-oriented tools remain present
    assert "hermes_session_send" in text
    assert "hermes_session_status" in text
    assert "hermes_sessions_list" in text
    assert "hermes_job_start" in text

    # Session send uses chat + bounded session interaction flags
    assert "chat" in text
    assert "-Q" in text
    assert "--continue" in text
    assert "--query" in text
    assert "sessions" in text
    assert "status" in text


def test_write_server_imports_write_helpers() -> None:
    text = read_text("tools/space-rocks-mcp/server-write.js")

    assert "repo_write_tools" in text
    assert "engineforge_write_tools" in text
    assert "restricted_command_tools" in text
    assert "registerRestrictedCommandTools" in text
    assert "job_tools" in text
    assert "registerJobTools" in text
    assert "defaultProcessJobManager" in text
    assert 'version: "0.3.0"' in text


def test_dev_server_registers_the_consolidated_development_surface() -> None:
    text = read_text("tools/space-rocks-mcp/server-dev.js")
    package_text = read_text("tools/space-rocks-mcp/package.json")

    assert 'name: "space-rocks-dev-mcp"' in text
    assert 'version: "0.1.0"' in text
    assert "process.env.PORT ?? 8889" in text
    assert '"start:dev": "PORT=8889 node server-dev.js"' in package_text

    assert "repo_readonly_tools" in text
    assert "repo_write_tools" in text
    assert 'if (name === "ping")' in text
    assert "restricted_command_tools" in text
    assert "registerRestrictedCommandTools" in text
    assert "job_tools" in text
    assert "registerJobTools" in text
    assert "defaultProcessJobManager" in text
    assert "registerHermesTools" in text
    assert "processJobManager: defaultProcessJobManager" in text
    assert "engineforge_readonly_tools" in text
    assert "engineforge_write_tools" in text
    assert "registerEngineForgeReadonlyTools" in text
    assert "registerEngineForgeWriteTools" in text

    assert "ENABLE_CHROME_DEVTOOLS" in text
    assert "registerChromeDevtoolsProxyTools" in text
    assert "registerPlasmicReadTools" in text
    assert "registerPlasmicWriteTools" in text


def test_repo_write_tools_use_workspace_write_service_and_batch_tool() -> None:
    text = read_text("tools/space-rocks-mcp/shared/repo_write_tools.js")

    assert "workspace_writes" in text
    assert "defaultWorkspaceWriteService" in text
    assert "apply_repo_file_edits" in text
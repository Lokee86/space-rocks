from pathlib import Path

from tools.architecture_guard import run_guard


def test_repository_architecture_rules() -> None:
    root = Path(__file__).resolve().parents[2]
    findings = run_guard(root)
    formatted = "\n".join(finding.format() for finding in findings)

    assert findings == [], f"Architecture guard findings:\n{formatted}"


def test_repository_rules_file_flags_gameplay_math_rand_but_not_rng_or_tests(tmp_path: Path) -> None:
    repo_root = Path(__file__).resolve().parents[2]
    config = tmp_path / "tools" / "architecture_guard" / "rules.toml"
    config.parent.mkdir(parents=True, exist_ok=True)
    config.write_text((repo_root / "tools" / "architecture_guard" / "rules.toml").read_text(encoding="utf-8"), encoding="utf-8")

    gameplay_file = tmp_path / "services/game-server/internal/game/spawning.go"
    gameplay_file.parent.mkdir(parents=True, exist_ok=True)
    gameplay_file.write_text("package game\n\nimport \"math/rand\"\n", encoding="utf-8")

    devtools_file = tmp_path / "services/game-server/internal/devtools/seed_debug.go"
    devtools_file.parent.mkdir(parents=True, exist_ok=True)
    devtools_file.write_text("package devtools\n\nimport \"math/rand\"\n", encoding="utf-8")

    rng_file = tmp_path / "services/game-server/internal/game/rng/seed.go"
    rng_file.parent.mkdir(parents=True, exist_ok=True)
    rng_file.write_text("package rng\n\nimport \"math/rand\"\n", encoding="utf-8")

    test_file = tmp_path / "services/game-server/internal/game/spawning_test.go"
    test_file.parent.mkdir(parents=True, exist_ok=True)
    test_file.write_text("package game\n\nimport \"math/rand\"\n", encoding="utf-8")

    findings = run_guard(tmp_path, config)

    assert {(finding.rule_id, finding.path, finding.line) for finding in findings} == {
        ("game-server-no-process-global-math-rand", "services/game-server/internal/game/spawning.go", 3),
        ("game-server-no-process-global-math-rand", "services/game-server/internal/devtools/seed_debug.go", 3),
    }


def test_repository_rules_block_plaintext_auth_token_persistence(tmp_path: Path) -> None:
    repo_root = Path(__file__).resolve().parents[2]
    config = tmp_path / "tools" / "architecture_guard" / "rules.toml"
    config.parent.mkdir(parents=True, exist_ok=True)
    config.write_text((repo_root / "tools" / "architecture_guard" / "rules.toml").read_text(encoding="utf-8"), encoding="utf-8")

    insecure_file = tmp_path / "client/scripts/auth/insecure_store.gd"
    insecure_file.parent.mkdir(parents=True, exist_ok=True)
    insecure_file.write_text(
        "extends RefCounted\n"
        "var path = \"user://auth_token.json\"\n"
        "func save(token: String) -> String:\n"
        "    return JSON.stringify({\"token\": token})\n",
        encoding="utf-8",
    )

    allowed_legacy_reader = tmp_path / "client/scripts/auth/auth_token_store.gd"
    allowed_legacy_reader.write_text(
        "extends RefCounted\n"
        "var token_path = \"user://auth_token.json\"\n"
        "func load_token() -> String:\n"
        "    return \"\"\n",
        encoding="utf-8",
    )

    findings = run_guard(tmp_path, config)

    assert {(finding.rule_id, finding.path, finding.line) for finding in findings} == {
        ("client-legacy-auth-token-path-is-read-only", "client/scripts/auth/insecure_store.gd", 2),
        ("client-no-json-bearer-token-persistence", "client/scripts/auth/insecure_store.gd", 4),
    }

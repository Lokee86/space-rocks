from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]


def test_client_fixture_mirror_matches_shared_fixture() -> None:
    source = REPO_ROOT / "shared/contracts/observability/fixtures/emitter_cases.json"
    mirror = REPO_ROOT / "client/tests/fixtures/observability_emitter_cases.json"
    assert mirror.read_bytes() == source.read_bytes()


def test_bridge_event_literal_is_confined_to_contract_generated_adapter_and_test_paths() -> None:
    runtime_roots = (
        REPO_ROOT / "shared/go",
        REPO_ROOT / "services",
        REPO_ROOT / "client/scripts",
    )
    allowed_parts = {
        "generated",
        "observabilityevent",
    }
    allowed_files = {
        REPO_ROOT / "client/scripts/logging/logger.gd",
        REPO_ROOT / "services/api-server/app/lib/observability/emitter.rb",
        REPO_ROOT / "services/api-server/app/lib/observability/structured_formatter.rb",
    }
    violations: list[str] = []
    for root in runtime_roots:
        for suffix in ("*.go", "*.rb", "*.gd"):
            for path in root.rglob(suffix):
                if "log_message" not in path.read_text(encoding="utf-8"):
                    continue
                if path.name.endswith("_test.go") or {"test", "tests"}.intersection(path.parts):
                    continue
                if "generated" in path.name or path in allowed_files or allowed_parts.intersection(path.parts):
                    continue
                violations.append(path.relative_to(REPO_ROOT).as_posix())
    assert violations == []


def test_service_adapters_use_generated_service_registry_constants() -> None:
    expectations = {
        "services/game-server/internal/logging/logger.go": (
            "observability.ServiceNameGameServer",
            "observability.ServiceKeyGameServer",
        ),
        "services/player-data/logging/logger.go": (
            "observability.ServiceNamePlayerData",
            "observability.ServiceKeyPlayerData",
        ),
        "services/diagnostic-aggregator/internal/logging/logger.go": (
            "observability.ServiceNameDiagnosticAggregator",
            "observability.ServiceKeyDiagnosticAggregator",
        ),
        "services/api-server/app/lib/observability/emitter.rb": (
            "ContractGenerated::SERVICE_API_SERVER",
        ),
        "client/scripts/logging/observability_emitter.gd": (
            "Contract.SERVICE_CLIENT",
        ),
    }
    for relative_path, generated_names in expectations.items():
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        for generated_name in generated_names:
            assert generated_name in content, f"{relative_path} must use {generated_name}"


def test_runtime_adapters_do_not_bypass_canonical_emitters() -> None:
    go_adapters = (
        "services/game-server/internal/logging/logger.go",
        "services/player-data/logging/logger.go",
        "services/diagnostic-aggregator/internal/logging/logger.go",
    )
    for relative_path in go_adapters:
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        assert ".Logger()." not in content
        assert "slog.NewJSONHandler" not in content

    client_adapter = (REPO_ROOT / "client/scripts/logging/logger.gd").read_text(encoding="utf-8")
    assert "_file_writer.write_line(" not in client_adapter

    for relative_path in (
        "services/api-server/app/lib/observability/structured_formatter.rb",
        "services/api-server/app/lib/observability/worker_runtime.rb",
    ):
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        assert "@writer.write(" not in content


def test_diagnostic_aggregator_has_no_production_bridge_calls() -> None:
    root = REPO_ROOT / "services/diagnostic-aggregator"
    forbidden = ("EmitLegacy(", "EmitLegacyArgs(")
    violations: list[str] = []
    for path in root.rglob("*.go"):
        if path.name.endswith("_test.go") or {".worktrees", ".workingtrees"}.intersection(path.parts):
            continue
        content = path.read_text(encoding="utf-8")
        for marker in forbidden:
            if marker in content:
                violations.append(
                    f"{path.relative_to(REPO_ROOT).as_posix()}: {marker}"
                )


    logger = (root / "internal/logging/logger.go").read_text(encoding="utf-8")
    assert "func (l *Logger) Info(" not in logger
    assert "func (l *Logger) Error(" not in logger

    hosted = (root / "hosted/service.go").read_text(encoding="utf-8")
    assert ".Info(" not in hosted and ".Error(" not in hosted
    assert violations == []

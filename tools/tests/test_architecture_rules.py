from pathlib import Path

from tools.architecture_guard import run_guard


def test_repository_architecture_rules() -> None:
    root = Path(__file__).resolve().parents[2]
    findings = run_guard(root)
    formatted = "\n".join(finding.format() for finding in findings)

    assert findings == [], f"Architecture guard findings:\n{formatted}"

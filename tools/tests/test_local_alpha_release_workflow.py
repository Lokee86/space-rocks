from pathlib import Path

import yaml


WORKFLOW = Path(__file__).resolve().parents[2] / ".github" / "workflows" / "local-alpha-release-gate.yml"


def test_local_alpha_workflow_publishes_native_release_archives() -> None:
    workflow = WORKFLOW.read_text(encoding="utf-8")
    parsed = yaml.safe_load(workflow)

    assert "publish-local-alpha" in parsed["jobs"]
    assert "Create Windows release archive" in workflow
    assert "Create macOS release archive" in workflow
    assert "ditto -c -k --sequesterRsrc" in workflow
    assert workflow.count("archive: false") == 2

    assert "publish-local-alpha:" in workflow
    assert "- windows-local-alpha" in workflow
    assert "- macos-local-alpha" in workflow
    assert "contents: write" in workflow
    assert "space-rocks-local-alpha-macos.zip" in workflow
    assert "macos-universal" not in workflow
    assert "--repo \"$GITHUB_REPOSITORY\"" in workflow
    assert "gh release create" in workflow
    assert "--prerelease --latest=false" in workflow

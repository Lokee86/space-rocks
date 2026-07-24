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
    assert "tools/release/version.txt" in workflow
    assert "space-rocks-${RELEASE_VERSION}-macos.zip" in workflow
    assert "space-rocks-${RELEASE_VERSION}-windows-x86_64.zip" in workflow
    assert "macos-universal" not in workflow
    assert "actions/runs/$GITHUB_RUN_ID/artifacts" in workflow
    assert "actions/artifacts/$windows_artifact_id/zip" in workflow
    assert "actions/artifacts/$macos_artifact_id/zip" in workflow
    assert "gh run download" not in workflow
    assert "unzip -tq" in workflow
    assert "startsWith(github.ref, 'refs/tags/v')" in workflow
    assert 'expected_tag="v$RELEASE_VERSION"' in workflow
    assert 'release_tag="$GITHUB_REF_NAME"' in workflow
    assert "--repo \"$GITHUB_REPOSITORY\"" in workflow
    assert "gh release create" in workflow
    assert "--prerelease --latest=false" in workflow

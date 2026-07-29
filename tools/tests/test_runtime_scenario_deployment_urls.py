from __future__ import annotations

import pytest

from runtime_scenarios.runner import deployment_server_urls


def test_https_deployment_url_uses_secure_websocket_and_health() -> None:
    websocket_url, health_url = deployment_server_urls("https://alpha.example.com")

    assert websocket_url == "wss://alpha.example.com/ws"
    assert health_url == "https://alpha.example.com/health"


def test_explicit_websocket_path_is_preserved() -> None:
    websocket_url, health_url = deployment_server_urls(
        "wss://alpha.example.com/game/ws?ticket=test"
    )

    assert websocket_url == "wss://alpha.example.com/game/ws?ticket=test"
    assert health_url == "https://alpha.example.com/health"


def test_deployment_url_requires_supported_scheme_and_host() -> None:
    with pytest.raises(ValueError, match="server URL"):
        deployment_server_urls("alpha.example.com")

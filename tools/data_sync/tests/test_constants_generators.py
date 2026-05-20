from __future__ import annotations

import pytest

from data_sync.generators import gds_constants, go_constants, ts_constants
from data_sync.generators.go_constants import ConstantsGenerationError


VALUES = (
    ("player_speed", 420.0),
    ("tick_rate", 60),
    ("debug_enabled", True),
    ("welcome_text", "hello"),
)


def test_go_constants_all_supported_value_types() -> None:
    assert go_constants.generate_constants("constants.gameplay", VALUES) == "\n".join(
        [
            "const PlayerSpeed = 420.0",
            "const TickRate = 60",
            "const DebugEnabled = true",
            'const WelcomeText = "hello"',
        ]
    )


def test_gds_constants_all_supported_value_types() -> None:
    assert gds_constants.generate_constants("constants.gameplay", VALUES) == "\n".join(
        [
            "const PLAYER_SPEED := 420.0",
            "const TICK_RATE := 60",
            "const DEBUG_ENABLED := true",
            'const WELCOME_TEXT := "hello"',
        ]
    )


def test_gds_constants_support_vector2_values() -> None:
    assert gds_constants.generate_constants(
        "constants.client.presentation",
        (("window_min_size", [1280.0, 720.0]),),
    ) == "const WINDOW_MIN_SIZE := Vector2(1280.0, 720.0)"


def test_ts_constants_all_supported_value_types() -> None:
    assert ts_constants.generate_constants("constants.gameplay", VALUES) == "\n".join(
        [
            "export const PLAYER_SPEED = 420.0;",
            "export const TICK_RATE = 60;",
            "export const DEBUG_ENABLED = true;",
            'export const WELCOME_TEXT = "hello";',
        ]
    )


def test_stable_output_order() -> None:
    values = (
        ("third_value", 3),
        ("first_value", 1),
        ("second_value", 2),
    )

    assert go_constants.generate_constants("constants.order", values).splitlines() == [
        "const ThirdValue = 3",
        "const FirstValue = 1",
        "const SecondValue = 2",
    ]


@pytest.mark.parametrize(
    "generator",
    [
        go_constants.generate_constants,
        gds_constants.generate_constants,
        ts_constants.generate_constants,
    ],
)
def test_invalid_constant_name_fails(generator) -> None:
    with pytest.raises(ConstantsGenerationError):
        generator("constants.bad", (("not-snake", 1),))


@pytest.mark.parametrize(
    "generator",
    [
        go_constants.generate_constants,
        gds_constants.generate_constants,
        ts_constants.generate_constants,
    ],
)
def test_unsupported_value_type_fails(generator) -> None:
    if generator is gds_constants.generate_constants:
        pytest.skip("GDScript supports Vector2 list values")
    with pytest.raises(ConstantsGenerationError):
        generator("constants.bad", (("vector_value", [1.0, 2.0]),))

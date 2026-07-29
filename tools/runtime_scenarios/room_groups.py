from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Callable

from runtime_scenarios.model import Scenario
from runtime_scenarios.processes import ManagedProcess


@dataclass
class RoomGroup:
    index: int
    coordinator: ManagedProcess
    room_code: str
    participants: list[ManagedProcess] = field(default_factory=list)

    @property
    def clients(self) -> list[ManagedProcess]:
        return [self.coordinator, *self.participants]

    def summary(self) -> dict[str, Any]:
        return {
            "index": self.index,
            "room_code": self.room_code,
            "coordinator": self.coordinator.name,
            "participants": [participant.name for participant in self.participants],
        }


def launch_room_groups(
    *,
    scenario: Scenario,
    godot: Path,
    headless_coordinator: bool,
    start_client: Callable[..., ManagedProcess],
    wait_for_status: Callable[..., dict[str, Any]],
) -> list[RoomGroup]:
    groups: list[RoomGroup] = []
    seen_codes: set[str] = set()
    for room_index in range(1, scenario.room_count + 1):
        coordinator = start_client(
            godot=godot,
            client_id=coordinator_client_id(scenario.room_count, room_index),
            role="coordinator",
            headless=headless_coordinator,
        )
        room_status = wait_for_status(
            coordinator,
            accepted={"room_ready"},
            timeout=scenario.setup_timeout_seconds,
        )
        room_code = str(room_status.get("room_code", "")).strip()
        if not room_code:
            raise RuntimeError(f"{coordinator.name} did not publish a room code")
        if room_code in seen_codes:
            raise RuntimeError(f"room code {room_code} was reused")
        seen_codes.add(room_code)
        groups.append(RoomGroup(room_index, coordinator, room_code))

    for group in groups:
        for participant_index in range(1, scenario.clients.headless + 1):
            group.participants.append(
                start_client(
                    godot=godot,
                    client_id=participant_client_id(
                        scenario.room_count,
                        group.index,
                        participant_index,
                    ),
                    role="participant",
                    headless=True,
                    room_code=group.room_code,
                )
            )
    return groups


def coordinator_client_id(room_count: int, room_index: int) -> str:
    if room_count == 1:
        return "coordinator-1"
    return f"room-{room_index}-coordinator"


def participant_client_id(
    room_count: int,
    room_index: int,
    participant_index: int,
) -> str:
    if room_count == 1:
        return f"participant-{participant_index}"
    return f"room-{room_index}-participant-{participant_index}"

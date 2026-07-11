from __future__ import annotations

import json

from data_sync.model.realtime_wire import RealtimeWireContract


def generate_realtime_wire_json(contract: RealtimeWireContract) -> str:
    return json.dumps(contract.as_dict(), indent=2, sort_keys=True) + "\n"

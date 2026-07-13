from __future__ import annotations

import json

from data_sync.model.observability import ObservabilityContract


def generate_observability_json(contract: ObservabilityContract) -> str:
    return json.dumps(
        contract.as_dict(),
        ensure_ascii=False,
        indent=2,
        sort_keys=True,
    ) + "\n"

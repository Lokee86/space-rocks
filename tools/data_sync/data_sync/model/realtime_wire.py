from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Mapping


ENCODING_MODES = frozenset(
    {
        "map",
        "fixed_tuple",
        "sparse_positional_tuple",
        "sparse_key_value_tuple",
        "scalar_id_list",
        "scalar_id",
        "scalar_list",
        "discriminated_event_tuple",
    }
)


@dataclass(frozen=True)
class RealtimeWireKeyAlias:
    domain: str
    readable: str
    compact: str


@dataclass(frozen=True)
class RealtimeWireValueAlias:
    readable: str
    compact: str


@dataclass(frozen=True)
class RealtimeWireValueDomain:
    id: str
    entries: tuple[RealtimeWireValueAlias, ...]


@dataclass(frozen=True)
class RealtimeWireIDCodec:
    id: str
    prefix: str
    tag: str | None = None
    numeric_suffix: bool = True
    preserve_malformed: bool = True


@dataclass(frozen=True)
class RealtimeWireIDSelectorMapping:
    value: str
    codec_id: str


@dataclass(frozen=True)
class RealtimeWireIDSelector:
    id: str
    field: str
    mappings: tuple[RealtimeWireIDSelectorMapping, ...]
    fallback_tagged: bool = False


@dataclass(frozen=True)
class RealtimeWireQuantization:
    path: str
    policy: str


@dataclass(frozen=True)
class RealtimeWirePacket:
    id: str
    compact: str
    lane: str
    snapshot_kind: str | None = None
    runtime: bool = False
    infer_lane: bool = False
    infer_snapshot_kind: bool = False
    infer_snapshot_id: bool = False
    infer_baseline_id: bool = False
    use_baseline_sequence: bool = False
    omit_single_chunk_metadata: bool = False


@dataclass(frozen=True)
class RealtimeWireField:
    name: str
    json: str
    quantization: str | None = None
    id_codec: str | None = None
    id_codec_by: str | None = None
    value_domain: str | None = None


@dataclass(frozen=True)
class RealtimeWireRecord:
    id: str
    source_struct: str | None
    encoding: str
    identity_field: str | None
    sparse_placeholder: str | None
    sparse_trailing: bool
    preserve_unknown_fields: bool = False
    fields: tuple[RealtimeWireField, ...] = ()


@dataclass(frozen=True)
class RealtimeWirePacketField:
    packet_id: str
    readable_field: str
    record_id: str
    decode_record_ids: tuple[str, ...] = ()


@dataclass(frozen=True)
class RealtimeWireEvent:
    readable: str
    compact: str
    record_id: str


@dataclass(frozen=True)
class RealtimeWireContract:
    id: str
    version: int
    readable_keys: bool
    explicit_metadata: bool
    unknown_key_passthrough: bool = False
    unknown_event_passthrough: bool = False
    key_aliases: tuple[RealtimeWireKeyAlias, ...] = ()
    value_domains: tuple[RealtimeWireValueDomain, ...] = ()
    id_codecs: tuple[RealtimeWireIDCodec, ...] = ()
    id_selectors: tuple[RealtimeWireIDSelector, ...] = ()
    quantizations: tuple[RealtimeWireQuantization, ...] = ()
    packets: tuple[RealtimeWirePacket, ...] = ()
    records: tuple[RealtimeWireRecord, ...] = ()
    packet_fields: tuple[RealtimeWirePacketField, ...] = ()
    events: tuple[RealtimeWireEvent, ...] = ()

    def value_domain(self, domain_id: str) -> RealtimeWireValueDomain:
        for domain in self.value_domains:
            if domain.id == domain_id:
                return domain
        raise KeyError(domain_id)

    def record(self, record_id: str) -> RealtimeWireRecord:
        for record in self.records:
            if record.id == record_id:
                return record
        raise KeyError(record_id)

    def packet(self, packet_id: str) -> RealtimeWirePacket:
        for packet in self.packets:
            if packet.id == packet_id:
                return packet
        raise KeyError(packet_id)

    def as_dict(self) -> Mapping[str, Any]:
        return {
            "wire_contract": {
                "id": self.id,
                "version": self.version,
                "readable_keys": self.readable_keys,
                "explicit_metadata": self.explicit_metadata,
                "unknown_key_passthrough": self.unknown_key_passthrough,
                "unknown_event_passthrough": self.unknown_event_passthrough,
            },
            "key_aliases": [alias.__dict__ for alias in self.key_aliases],
            "value_domains": [
                {"id": domain.id, "entries": [entry.__dict__ for entry in domain.entries]}
                for domain in self.value_domains
            ],
            "id_codecs": [codec.__dict__ for codec in self.id_codecs],
            "id_selectors": [
                {
                    "id": selector.id,
                    "field": selector.field,
                    "fallback_tagged": selector.fallback_tagged,
                    "mappings": [mapping.__dict__ for mapping in selector.mappings],
                }
                for selector in self.id_selectors
            ],
            "quantizations": [quantization.__dict__ for quantization in self.quantizations],
            "packets": [packet.__dict__ for packet in self.packets],
            "records": [
                {
                    **{key: value for key, value in record.__dict__.items() if key != "fields"},
                    "fields": [field.__dict__ for field in record.fields],
                }
                for record in self.records
            ],
            "packet_fields": [
                {
                    "packet_id": binding.packet_id,
                    "readable_field": binding.readable_field,
                    "record_id": binding.record_id,
                    "decode_record_ids": list(binding.decode_record_ids),
                }
                for binding in self.packet_fields
            ],
            "events": [event.__dict__ for event in self.events],
        }

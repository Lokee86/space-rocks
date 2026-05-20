"""Generated block extraction and replacement."""

from __future__ import annotations

import re
from dataclasses import dataclass


MARKER_RE = re.compile(
    r"^(?P<indent>[ \t]*)(?P<comment>//|#)[ \t]*data-sync:(?P<kind>start|end)[ \t]+(?P<section>\S+)[ \t]*$",
    re.MULTILINE,
)


class BlockIOError(Exception):
    """Raised when generated block markers are missing or ambiguous."""


@dataclass(frozen=True)
class ManagedBlock:
    section: str
    start_marker_start: int
    start_marker_end: int
    content_start: int
    content_end: int
    end_marker_start: int
    end_marker_end: int
    comment: str

    @property
    def content(self) -> str:
        raise AttributeError("ManagedBlock does not store source text; use extract_block")


def find_block(text: str, section: str) -> ManagedBlock:
    blocks = [block for block in find_all_blocks(text) if block.section == section]
    if not blocks:
        starts = _marker_sections(text, "start")
        ends = _marker_sections(text, "end")
        if section in starts and section not in ends:
            raise BlockIOError(f"missing data-sync end marker for section {section!r}")
        if section in ends and section not in starts:
            raise BlockIOError(f"missing data-sync start marker for section {section!r}")
        raise BlockIOError(f"missing data-sync block for section {section!r}")
    if len(blocks) > 1:
        raise BlockIOError(f"duplicate data-sync block for section {section!r}")
    return blocks[0]


def extract_block(text: str, section: str) -> str:
    block = find_block(text, section)
    return text[block.content_start : block.content_end]


def replace_block(text: str, section: str, new_content: str) -> str:
    block = find_block(text, section)
    replacement = _canonical_block_content(new_content)
    return text[: block.content_start] + replacement + text[block.content_end :]


def find_all_blocks(text: str) -> tuple[ManagedBlock, ...]:
    markers = list(_iter_markers(text))
    blocks: list[ManagedBlock] = []
    open_starts: dict[str, _Marker] = {}
    completed_sections: set[str] = set()

    for marker in markers:
        if marker.kind == "start":
            if marker.section in open_starts or marker.section in completed_sections:
                raise BlockIOError(f"duplicate data-sync block for section {marker.section!r}")
            open_starts[marker.section] = marker
            continue

        start_marker = open_starts.pop(marker.section, None)
        if start_marker is None:
            raise BlockIOError(f"missing data-sync start marker for section {marker.section!r}")
        if start_marker.comment != marker.comment:
            raise BlockIOError(
                f"data-sync marker comment style mismatch for section {marker.section!r}"
            )

        completed_sections.add(marker.section)
        blocks.append(
            ManagedBlock(
                section=marker.section,
                start_marker_start=start_marker.start,
                start_marker_end=start_marker.end,
                content_start=start_marker.end,
                content_end=marker.start,
                end_marker_start=marker.start,
                end_marker_end=marker.end,
                comment=marker.comment,
            )
        )

    if open_starts:
        section = next(iter(open_starts))
        raise BlockIOError(f"missing data-sync end marker for section {section!r}")

    return tuple(blocks)


@dataclass(frozen=True)
class _Marker:
    section: str
    kind: str
    start: int
    end: int
    comment: str


def _iter_markers(text: str) -> tuple[_Marker, ...]:
    markers: list[_Marker] = []
    for match in MARKER_RE.finditer(text):
        end = match.end()
        if end < len(text) and text[end : end + 1] == "\n":
            end += 1
        markers.append(
            _Marker(
                section=match.group("section"),
                kind=match.group("kind"),
                start=match.start(),
                end=end,
                comment=match.group("comment"),
            )
        )
    return tuple(markers)


def _marker_sections(text: str, kind: str) -> set[str]:
    return {marker.section for marker in _iter_markers(text) if marker.kind == kind}


def _canonical_block_content(new_content: str) -> str:
    content = new_content.strip("\n")
    if not content:
        return ""
    return f"{content}\n"

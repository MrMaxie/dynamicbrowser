#!/usr/bin/env python3
"""Check the stable release contract and extract notes without publishing."""

import argparse
import re
import subprocess
import sys
from datetime import date
from pathlib import Path

VERSION = r"(?:0|[1-9][0-9]*)(?:\.(?:0|[1-9][0-9]*)){2}"
RELEASE = re.compile(rf"## \[({VERSION})\] - ([0-9]{{4}}-[0-9]{{2}}-[0-9]{{2}})")


def version_key(value):
    if not re.fullmatch(VERSION, value):
        raise ValueError(f"expected a stable SemVer version, got {value!r}")
    return tuple(map(int, value.split(".")))


def releases(changelog):
    # shortcut: level-two headings are reserved for releases; use level three inside notes.
    sections = re.split(r"(?m)(?=^## )", changelog)
    if len(sections) < 2 or sections[1].splitlines()[0] != "## [Unreleased]":
        raise ValueError("CHANGELOG.md must start with a ## [Unreleased] section")
    result = {}
    for section in sections[2:]:
        heading, _, notes = section.partition("\n")
        match = RELEASE.fullmatch(heading)
        if not match:
            raise ValueError(f"invalid release heading: {heading!r}")
        version, day = match.groups()
        date.fromisoformat(day)
        if result and version_key(version) >= version_key(next(reversed(result))):
            raise ValueError("changelog releases must be unique and ordered newest first")
        visible = re.sub(r"<!--.*?-->", "", notes, flags=re.DOTALL)
        if not any(line.strip() and not line.lstrip().startswith("#") for line in visible.splitlines()):
            raise ValueError(f"changelog section {version} has no release notes")
        result[version] = (heading, notes.strip())
    return result


def check_pr(old_version, version, old_releases, new_releases):
    old, new = version_key(old_version), version_key(version)
    if new < old:
        raise ValueError("VERSION must not decrease")
    expected = list(old_releases) if new == old else [version, *old_releases]
    if list(new_releases) != expected:
        raise ValueError("a VERSION bump requires exactly one matching new changelog section; otherwise change Unreleased only")
    for released, section in old_releases.items():
        if new_releases[released] != section:
            raise ValueError(f"published changelog section {released} must not change")


def release_notes(tag, version, sections):
    version_key(version)
    if tag != f"v{version}" or next(iter(sections), None) != version:
        raise ValueError("tag, VERSION and the newest changelog release must match")
    return sections[version][1]


def git(*args):
    return subprocess.check_output(["git", *args], text=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--base", help="compare the working tree with this base revision")
    mode.add_argument("--tag", help="validate a tag and print only its release notes")
    args = parser.parse_args()
    try:
        version = Path("VERSION").read_text(encoding="utf-8").strip()
        sections = releases(Path("CHANGELOG.md").read_text(encoding="utf-8"))
        if args.base:
            base = git("rev-parse", "--verify", "--end-of-options", args.base + "^{commit}").strip()
            old_sections = releases(git("show", f"{base}:CHANGELOG.md")) if git("ls-tree", base, "--", "CHANGELOG.md") else {}
            check_pr(git("show", f"{base}:VERSION").strip(), version, old_sections, sections)
        else:
            print(release_notes(args.tag, version, sections))
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        print(f"Release contract: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())

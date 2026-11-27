import runpy
import unittest
from pathlib import Path

contract = runpy.run_path(str(Path(__file__).with_name("release-contract.py")))
releases = contract["releases"]
check_pr = contract["check_pr"]
release_notes = contract["release_notes"]

PREAMBLE = "# Changelog\n\n## [Unreleased]\n\n"
FIRST = "## [1.0.0] - 2026-11-27\n\n### Added\n\n- Linux builds.\n"
SECOND = "## [1.0.1] - 2026-12-01\n\n### Fixed\n\n- Tray behavior.\n"


class ReleaseContractTests(unittest.TestCase):
    def test_pr_can_add_pending_notes_without_bump(self):
        check_pr("1.0.0", "1.0.0", releases(PREAMBLE + FIRST),
                 releases(PREAMBLE + "- Pending fix.\n" + FIRST))

    def test_bump_requires_one_new_section(self):
        old = releases(PREAMBLE + FIRST)
        new = releases(PREAMBLE + SECOND + FIRST)
        check_pr("1.0.0", "1.0.1", old, new)
        check_pr("0.9.1", "1.0.0", {}, old)
        for before, after, base, proposed in (
            ("1.0.0", "1.0.1", old, old),
            ("1.0.0", "1.0.0", old, new),
            ("1.0.1", "1.0.0", new, old),
            ("1.0.0", "1.0.1", old, releases(PREAMBLE + SECOND)),
            ("1.0.0", "1.0.10", old, new),
        ):
            with self.subTest(before=before, after=after, proposed=list(proposed)), self.assertRaises(ValueError):
                check_pr(before, after, base, proposed)

    def test_invalid_release_sections(self):
        for body in (FIRST + FIRST, FIRST + SECOND,
                     FIRST.replace("2026-11-27", "2026-02-30"),
                     FIRST.replace("- Linux builds.", "<!-- no notes -->")):
            with self.subTest(body=body), self.assertRaises(ValueError):
                releases(PREAMBLE + body)

    def test_tag_extracts_only_matching_notes(self):
        sections = releases(PREAMBLE + "- Pending.\n" + SECOND + FIRST)
        self.assertEqual(release_notes("v1.0.1", "1.0.1", sections),
                         "### Fixed\n\n- Tray behavior.")
        with self.assertRaises(ValueError):
            release_notes("v1.0.0", "1.0.1", sections)
        with self.assertRaises(ValueError):
            release_notes("v1.0.0", "1.0.0", sections)

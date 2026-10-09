# Tasks

- [x] Implement typed YAML configuration, schema validation, ordered browsers, patterns, and last-valid-snapshot reloads.
- [x] Implement ordered rule selection, explicit fallback, direct executable arguments, and on-demand routing with resident handoff.
- [x] Implement the singleton tray, forced replacement, session-only selection, configuration editing, and user-level autostart.
- [x] Implement explicit Windows registration and restoration with separate persistent HTTP/HTTPS choices, supported system consent, and verified cleanup.
- [x] Verify normal registration, system-shell URL activation, repeated setup, cancellation, restoration, and mixed-protocol side-effect detection in Windows Sandbox.
- [x] Confirm initial live Windows routing through user acceptance; extended daily-use evaluation remains in progress.
- [x] Add a repeatable offline Sandbox harness, native test coverage, a routing probe, and a separate consent-dependent checklist.
- [x] Update README and validate the public example against the runtime schema and decoder.
- [x] Run final `just check` and independent review of documentation, harness, packaging, and privacy boundaries.
- [x] Review `release.md` against delivered outcomes.

Windows is the native verification target. Automated Sandbox results distinguish completed tests from manual Settings checks; they do not claim the operating-system consent flow was automated. Local ZIP preparation uses the current `VERSION` and does not publish or assign a new release version.

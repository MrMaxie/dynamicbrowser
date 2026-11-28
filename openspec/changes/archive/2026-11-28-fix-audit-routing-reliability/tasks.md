## Implementation and verification

- [x] Fix Linux runtime paths, tray readiness, sign-in setup and registration recovery.
- [x] Fix Windows source defaults, tray reload and concurrent handoff.
- [x] Drop test-only configuration state and prevent rewriting published changelog entries.
- [x] Run regression tests in an isolated Linux namespace with a private session bus for the headless tray test.
- [x] Run the automated Windows tray tests in GitHub Actions.

## Post-release verification (v1.0.1)

Passed, as reported by the user:

- [x] Windows Sandbox: set up HTTP and HTTPS handling, open links from another app, and restore the previous associations.
- [x] KDE Plasma/Wayland VM: open the tray, select a profile, and verify that links still open after closing the tray.

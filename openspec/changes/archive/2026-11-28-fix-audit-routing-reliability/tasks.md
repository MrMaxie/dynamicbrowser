## Implementation and verification

- [x] Fix Linux runtime paths, tray readiness, sign-in setup and registration recovery.
- [x] Fix Windows source defaults, tray reload and concurrent handoff.
- [x] Drop test-only configuration state and prevent rewriting published changelog entries.
- [x] Run regression tests in an isolated Linux namespace with a private session bus for the headless tray test.
- [x] Run the automated Windows tray tests in GitHub Actions.

Manual Windows Settings consent and a live Linux desktop were not checked for this release.

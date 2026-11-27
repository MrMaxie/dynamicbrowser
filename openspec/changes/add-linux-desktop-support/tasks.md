## 1. Shared routing and Linux desktop integration

- [ ] 1.1 Add the Linux XDG configuration path; keep the Windows path, schema, and executable resolution compatible. Test unset, relative, and absolute XDG values, plus tray and on-demand use of the same file from a different working directory.
- [ ] 1.2 Add Linux registration and restoration for HTTP and HTTPS. Test distinct previous handlers, declined consent, repeated setup, missing previous handlers, partial updates, retries, and an independently changed default. Keep existing Windows regression tests.
- [ ] 1.3 Verify Linux on-demand routing, unknown source metadata, tray handoff and override, live reload, editor launch, singleton replacement, no-host failure, and user-level autostart without making the tray a routing prerequisite.

## 2. Native builds and acceptance

- [ ] 2.1 Add native Linux build, check, and packaging commands alongside Windows. Preserve the Windows artifact contract and add an independent Linux Actions check with a versioned tarball and SHA-256 checksum.
- [ ] 2.2 On each native runner verify the archive contents, checksum, executable version, and URL launch into a test browser command. Run the existing regression suite on Windows and Linux.
- [ ] 2.3 In a disposable KDE Plasma Wayland session, install the Linux archive as a regular user, register both schemes, open links through the desktop, and restore distinct previous handlers without replacing a newer independent choice. Recheck Windows Settings registration and restoration with the existing Sandbox checklist.
- [ ] 2.4 In that Linux desktop session, verify the visible tray and native menu, profile override for desktop links, `Auto`, editing and live reload, singleton replacement with `--force`, and opt-in sign-in startup. Close the tray and verify that desktop links still route without relaunching it.
- [ ] 2.5 Document both targets in English, including install, configuration locations, system setup and undo, native artifact names, verified Linux desktop prerequisites, and the source-metadata limitation.

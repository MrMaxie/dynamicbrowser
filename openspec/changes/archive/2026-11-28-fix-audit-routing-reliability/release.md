---
impact: patch
visibility: public
components:
  - browser-routing
  - tray-routing
  - browser-registration
---

## Fixed

- On Linux, links open with long cache paths and when no tray host is available. An attempt to start a tray without a supported host now explains why it cannot open. Sign-in setup warnings no longer mark a link that opened as failed.
- On both platforms, the tray keeps the latest valid configuration even after a bad edit, and a slow browser launch no longer blocks all other links.
- On Linux, `--restore` can remove registration after both defaults have been changed manually, even if the recovery file is corrupt.

## Changed

- On Windows, source rules no longer guess the originating app from the focused window by default. Set `guess-windows-source: true` to keep the old behavior if you accept that a focus change can select the wrong profile. URL rules and tray overrides are unchanged.

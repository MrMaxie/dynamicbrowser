---
impact: minor
visibility: public
components:
  - desktop-targets
  - browser-registration
---

## Added

### Linux desktop support

Use dynamicbrowser on Linux amd64 to route HTTP and HTTPS links to browser profiles, with per-user default-browser setup and the same tray controls as on Windows on compatible desktops. Links still route when the tray is closed. Linux and Windows builds use the same version and configuration format; existing Windows setups are unchanged.

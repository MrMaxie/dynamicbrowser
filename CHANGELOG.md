# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/2.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2026-11-28

### Fixed

- On Linux, links open with long cache paths and when no tray host is available. An attempt to start a tray without a supported host now explains why it cannot open. Sign-in setup warnings no longer mark a link that opened as failed.
- On both platforms, the tray keeps the latest valid configuration even after a bad edit, and a slow browser launch no longer blocks all other links.
- On Linux, `--restore` can remove registration after both defaults have been changed manually, even if the recovery file is corrupt.

### Changed

- On Windows, source rules no longer guess the originating app from the focused window by default. Set `guess-windows-source: true` to keep the old behavior if you accept that a focus change can select the wrong profile. URL rules and tray overrides are unchanged.

## [1.0.0] - 2026-11-27

### Added

- Linux amd64 builds can route links to browser profiles with or without a tray. On compatible desktops, the tray offers the same routing controls as on Windows. Linux builds are available as a separate CI archive with a checksum. <!-- openspec: add-linux-desktop-support -->
- On Linux, `--register` sets up dynamicbrowser for HTTP and HTTPS without administrator access. `--restore` reinstates the previous handlers without replacing a newer choice. <!-- openspec: add-linux-desktop-support -->
- On Linux, source rules can match the launcher process when it is known. On Hyprland, window rules can use the active window class as a last resort; this fallback can be disabled. <!-- openspec: add-linux-desktop-support -->

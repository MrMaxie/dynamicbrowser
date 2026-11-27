# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/2.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-11-27

### Added

- Linux amd64 builds can route links to browser profiles with or without a tray. On compatible desktops, the tray offers the same routing controls as on Windows. Linux builds are available as a separate CI archive with a checksum. <!-- openspec: add-linux-desktop-support -->
- On Linux, `--register` sets up dynamicbrowser for HTTP and HTTPS without administrator access. `--restore` reinstates the previous handlers without replacing a newer choice. <!-- openspec: add-linux-desktop-support -->
- On Linux, source rules can match the launcher process when it is known. On Hyprland, window rules can use the active window class as a last resort; this fallback can be disabled. <!-- openspec: add-linux-desktop-support -->

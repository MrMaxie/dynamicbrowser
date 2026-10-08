# Why

Users need one default web-link handler that selects the appropriate browser profile from the URL and its source application, while allowing a temporary manual choice and a reversible setup.

# What changes

- Define named browser configurations with `exe` and an optional string-array `args`.
- Match process, window, full URL, domain and path using scalar glob strings or `regex:` strings; arrays provide OR alternatives and specified fields combine with AND.
- Use top-level `default` only when a rule cannot select an available browser, and show a system error if that fallback is unavailable.
- Route URLs on demand without a tray, or hand them to the resident tray to apply its session-only override.
- Provide the menu `Auto`, browser names, `Edit configuration`, and `Close tray`, and let `autostart` control tray startup at sign-in.
- Assist default-browser registration and restoration of the previous HTTP/HTTPS handlers through supported system mechanisms.
- Replace the initial greeting requirement with resident tray startup and retain the version command.
- Read executable-adjacent `config.yaml` and provide a local, extensible JSON Schema association for YAML language servers.

The requirements describe the intended product behavior, not completed functionality. Routing, tray interaction and default-browser setup can be implemented and verified in separate increments.

# Out of scope

- A graphical configuration editor, browser extensions, telemetry or a separate browser engine.
- Compatibility with another application's configuration format.
- A frozen configuration schema or workstation-specific executable paths.
- Bypassing operating-system consent or modifying protected default-association data directly.

# Capabilities

## New capabilities

- `browser-routing`: Named profiles, patterns, on-demand link delivery and live configuration.
- `tray-routing`: Resident tray startup and temporary profile selection.
- `browser-registration`: Assisted registration and reversible default-browser setup.

## Modified capabilities

- `application-cli`: Remove the greeting and preserve build-time version reporting without resident startup.

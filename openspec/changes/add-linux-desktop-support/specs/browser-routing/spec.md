## MODIFIED Requirements

### Requirement: Configuration format

The application SHALL read one YAML mapping from `config.yaml` independently of the working directory. On Windows the file SHALL remain next to the executable. On Linux it SHALL be under `$XDG_CONFIG_HOME/dynamicbrowser/`, defaulting to `$HOME/.config/dynamicbrowser/` when the XDG setting is unset or invalid. The local schema SHALL be adjacent to the configuration. Both platforms SHALL support optional top-level `default`, `autostart`, `browsers` and `rules` fields. `default` SHALL be a browser configuration name. Each named browser SHALL define `exe` and MAY define `args` as an array of strings. Executable paths containing directory components SHALL resolve relative to the application executable's directory when not absolute; bare names MAY also resolve through `PATH`. Browser arguments and destination URLs SHALL be passed directly to the executable without shell interpretation. A missing configuration SHALL be created without inventing executable paths.

#### Scenario: Launch a browser profile

- **GIVEN** a named browser defines `exe` and `args: ['--profile-directory=ExampleProfile']`
- **WHEN** that browser is selected for a URL
- **THEN** it receives the configured arguments followed by the URL as a separate argument

#### Scenario: Use per-user configuration with a read-only Linux installation

- **GIVEN** the Linux executable is in a read-only installation directory
- **WHEN** the application starts with a writable XDG configuration directory and no configuration yet
- **THEN** it creates `config.yaml` and `config.schema.json` in the user configuration directory
- **AND** it does not write next to the executable

## ADDED Requirements

### Requirement: Unavailable source metadata on Linux

Linux routing SHALL treat source metadata not supplied reliably by desktop activation as empty strings, using the shared rule semantics. It SHALL NOT use the router's own process or whichever window happens to have focus as a substitute. URL rules, fallback selection, and tray overrides SHALL work without source metadata.

#### Scenario: Open a desktop link without source metadata

- **GIVEN** a rule requires a non-empty source process and a later rule matches the URL domain
- **WHEN** Linux desktop activation supplies only the URL
- **THEN** the source-specific rule does not match
- **AND** the domain rule can route the link

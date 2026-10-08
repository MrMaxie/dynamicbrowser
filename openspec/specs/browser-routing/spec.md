# browser-routing Specification

## Purpose

Define YAML configuration, source/URL matching, profile selection and live reload for HTTP/HTTPS routing.

## Requirements

### Requirement: Configuration format

The application SHALL read one YAML mapping from `config.yaml` next to its executable, independently of the working directory. Configuration SHALL support optional top-level `default`, `autostart`, `browsers` and `rules` fields. `default` SHALL be a browser configuration name. Each named browser SHALL define `exe` and MAY define `args` as an array of strings. Browser arguments and each destination URL SHALL be passed directly to the executable without shell interpretation. A missing configuration SHALL be created without inventing executable paths.

#### Scenario: Launch a browser profile

- **GIVEN** a named browser defines `exe` and `args: ['--profile-directory=ExampleProfile']`
- **WHEN** that browser is selected for a URL
- **THEN** it receives the configured arguments followed by the URL as a separate argument

### Requirement: Patterns

A pattern SHALL be a string. Strings beginning with `regex:` SHALL use Go regular expressions after removing that prefix. Other strings SHALL use whole-value glob matching: `*` matches any sequence and `?` matches one character, including separators. Backslash SHALL escape literal wildcard characters. A pattern field MAY instead contain an array of strings, with alternatives joined by OR. An empty array SHALL not match. Invalid patterns SHALL report a configuration error.

#### Scenario: Match alternatives

- **GIVEN** a process condition contains `['Editor*', 'regex:^Viewer[0-9]+$']`
- **WHEN** the process is `EditorApp` or `Viewer2`
- **THEN** the condition matches
- **AND** the glob `Editor\*` matches a literal `Editor*`, not `EditorApp`

### Requirement: Rule conditions

Rules SHALL be evaluated in order, with the first matching rule selecting its `browser`. A rule MAY have `source.process`, `source.window`, `target.url`, `target.domain` and `target.path` pattern fields. Source values SHALL describe the originating process name and window title separately. Target values SHALL describe the full URL, hostname without port, and parsed URL path. All specified fields SHALL match (AND), while alternatives within one field use OR. Unknown source values SHALL be treated as empty strings. A rule without conditions SHALL match any valid web URL.

#### Scenario: Combine source and destination conditions

- **GIVEN** a rule specifies a process pattern, a window pattern, a domain pattern and a path pattern
- **WHEN** all four fields match
- **THEN** the rule selects its browser
- **WHEN** any field does not match
- **THEN** evaluation continues to the next rule

### Requirement: Default selection only when needed

If no rule matches, or the first matching rule names a missing or unavailable browser, routing SHALL use the browser named by top-level `default`. If that default is empty, unknown or unavailable, the application SHALL display an operating-system error dialog and SHALL NOT guess another browser. Missing or invalid `default` SHALL NOT prevent a valid matching rule or a valid tray override from opening a URL.

#### Scenario: Open a matched URL without a default

- **GIVEN** no valid default is configured
- **WHEN** a rule selects an available browser
- **THEN** that browser opens the URL without a default-related error

#### Scenario: Fail when fallback is needed

- **GIVEN** `default` is empty or names an unavailable browser
- **WHEN** no rule matches, or a matching rule names an unavailable browser
- **THEN** a system error dialog asks the user to configure a valid default
- **AND** no browser is selected by declaration order

### Requirement: Routing without a resident tray

The application SHALL accept one or more HTTP/HTTPS URLs per invocation and route them without requiring a resident tray. If the tray is running, requests SHALL be handed to it so its temporary override applies. If it is not running, the invocation SHALL load configuration, open the URLs and exit without opening a tray. A failed handoff SHALL not silently discard URLs. Closing the tray SHALL not disable later on-demand routing.

#### Scenario: Route after closing the tray

- **WHEN** the tray has been closed and the application is invoked with a URL
- **THEN** the URL is routed using the configuration
- **AND** no resident tray is started

### Requirement: Live configuration and editor schema

A resident tray SHALL watch `config.yaml`, including atomic replacement. Known configuration fields SHALL follow the supplied schema's types; null pattern conditions and unsupported nested fields SHALL be rejected rather than silently ignored. Additional top-level fields MAY be retained as extensions. Invalid YAML, field shapes or patterns SHALL retain the last valid configuration. An unavailable browser reference SHALL be handled by routing fallback, not reject the entire configuration. The application SHALL supply a local `config.schema.json` and a YAML language-server modeline in newly created configurations. Existing configuration and schema files SHALL be preserved at startup.

#### Scenario: Reload without restarting

- **WHEN** a valid configuration replaces the current file
- **THEN** subsequent requests and the browser menu use it without restarting
- **WHEN** a later edit has invalid YAML or a malformed regex
- **THEN** the last valid configuration remains usable

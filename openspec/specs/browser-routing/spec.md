# browser-routing Specification

## Purpose

Define YAML configuration, source/URL matching, profile selection and live reload for HTTP/HTTPS routing.
## Requirements
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

### Requirement: Patterns

A pattern SHALL be a string. Strings beginning with `regex:` SHALL use Go regular expressions after removing that prefix. Other strings SHALL use whole-value glob matching: `*` matches any sequence and `?` matches one character, including separators. Backslash SHALL escape literal wildcard characters. A pattern field MAY instead contain an array of strings, with alternatives joined by OR. An empty array SHALL not match. Invalid patterns SHALL report a configuration error.

#### Scenario: Match alternatives

- **GIVEN** a process condition contains `['Editor*', 'regex:^Viewer[0-9]+$']`
- **WHEN** the process is `EditorApp` or `Viewer2`
- **THEN** the condition matches
- **AND** the glob `Editor\*` matches a literal `Editor*`, not `EditorApp`

### Requirement: Rule conditions

Rules SHALL be evaluated in order, with the first matching rule selecting its `browser`. A rule MAY have `source.process`, `source.window`, `target.url`, `target.domain` and `target.path` pattern fields. Source values SHALL follow the platform-specific rules below. Target values SHALL describe the full URL, hostname without port, and parsed URL path. All specified fields SHALL match (AND), while alternatives within one field use OR. Unknown source values SHALL NOT match a source condition. A rule without conditions SHALL match any valid web URL.

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

### Requirement: Unavailable source metadata on Linux

Linux routing SHALL use the direct launcher executable name for `source.process` when available and not a recognized intermediary. Known desktop brokers, openers, and shells SHALL NOT stand in for the originating application. A source condition SHALL NOT match an unknown value, even with a wildcard. The same source matching policy SHALL apply with and without a resident tray. Linux SHALL support the optional boolean configuration keys `window-as-process` and `active-window-fallback`, each defaulting to `true`. Windows SHALL ignore both options.

When `window-as-process` is enabled, ordinary `source.window` conditions SHALL match the known launcher process name on Linux, not a window title. When `active-window-fallback` is enabled and the launcher is unknown, the router SHALL try the active window's class on Hyprland only if no ordinary rule matches and no tray override applies. The class SHALL be used only to match `source.window`, not to supply `source.process`. If the class cannot be obtained or no rule matches it, the configured default SHALL apply. Disabling either option SHALL NOT disable the other.

#### Scenario: Open a desktop link without source metadata

- **GIVEN** a rule requires a non-empty source process and a later rule matches the URL domain
- **WHEN** Linux desktop activation supplies only the URL
- **THEN** the source-specific rule does not match
- **AND** the domain rule can route the link

#### Scenario: Ignore a known intermediary

- **GIVEN** a source rule uses the `*` process pattern
- **WHEN** Linux can identify only an opener or desktop broker
- **THEN** the source rule does not match, whether routing is on demand or handed to the tray

#### Scenario: Use the active class only as a last resort

- **GIVEN** a `source.window` rule matches the active Hyprland window class
- **WHEN** the launcher is unknown and no ordinary rule matches
- **THEN** the window rule can select the browser unless `active-window-fallback` is `false`
- **AND** an ordinary URL rule takes precedence whenever it matches

### Requirement: Opt-in routing diagnostics

When `DYNAMICBROWSER_TRACE_ROUTING=1` is set on the process performing routing, it SHALL write the known process name or `unknown`, any active window class consulted, the decision (rule number, default, override, or rejection), and the selected browser label to standard error. The trace SHALL be off by default and SHALL NOT include URLs, window titles, or browser arguments. For a request handed to a resident tray, tracing is controlled by the tray's environment.

#### Scenario: Diagnose a forwarded link

- **GIVEN** routing diagnostics are enabled on a resident tray
- **WHEN** a request is forwarded to it and a rule selects a browser
- **THEN** the tray reports the process, rule number, and browser label without logging the URL

### Requirement: Windows source guesses are opt-in

For URL launches that do not identify the originating app on Windows, DynamicBrowser SHALL treat the source process and window as unknown unless `guess-windows-source` is true. Unknown source fields SHALL NOT match a configured source condition, including a wildcard. When the option is true, a known process name and title from the focused window MAY be used separately as guesses.

#### Scenario: Focus changes before routing

- **GIVEN** a source-only work rule and a different default profile
- **WHEN** a link arrives without source metadata while the work window is focused
- **THEN** the work rule does not match just because of focus, unless the user enabled the guess

### Requirement: Tray and sign-in failures are separate from URL delivery

On Linux, link routing SHALL work without a supported tray host and SHALL NOT depend on the length of the cache directory. A tray launch without a supported host SHALL report that tray mode is unavailable and exit. Direct link routing SHALL report sign-in setup errors separately from the success or failure of handing the link to a browser. Tray handoff SHALL allow concurrent requests within a fixed bound and explicitly reject new work when all slots are occupied.

#### Scenario: No Linux tray host and a long cache path

- **GIVEN** no supported tray host and a long cache path
- **WHEN** a URL arrives after a failed attempt to start the tray
- **THEN** the configured browser receives the URL without starting a tray

#### Scenario: An invalid edit follows a valid one before the tray can update

- **GIVEN** the watcher has already read a valid edit followed by an invalid one
- **WHEN** the tray next processes the change notice
- **THEN** it uses the newest valid settings and reports the invalid edit

#### Scenario: A slow request is in progress

- **WHEN** a second link arrives while another request is slow and a slot is free
- **THEN** the second request can proceed without waiting for that first request

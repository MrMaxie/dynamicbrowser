## ADDED Requirements

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

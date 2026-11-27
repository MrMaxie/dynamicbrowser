## MODIFIED Requirements

### Requirement: Assisted default-browser registration

`--register` SHALL explicitly register the application as an available HTTP/HTTPS handler and assist the user in selecting it through supported operating-system mechanisms on Windows and Linux. Startup and link handling SHALL NOT change defaults. Platform integration SHALL be isolated and SHALL NOT bypass protected associations. Unsupported environments SHALL report the limitation without claiming success. Registration SHALL report success only after verifying both effective associations.

#### Scenario: Choose the application as the default browser on Windows

- **WHEN** the user requests registration on Windows
- **THEN** the application becomes available as a browser choice
- **AND** the supported default-app settings flow is opened when user confirmation is required
- **AND** success is reported only after the HTTP/HTTPS associations are verified

#### Scenario: Choose the application as the default browser on Linux

- **WHEN** the user requests registration in a supported Linux session
- **THEN** a user-level desktop entry advertises the HTTP and HTTPS handlers without requiring elevated privileges
- **AND** only explicit user consent authorizes changing the defaults
- **AND** both schemes are verified separately before success is reported

### Requirement: Restore previous defaults

Before assisting a change to the default browser, the application SHALL retain the previous effective HTTP and HTTPS handlers separately in persistent per-user state. Repeated registration SHALL NOT replace that snapshot with the application's own associations. `--restore` SHALL restore only associations still owned by dynamicbrowser, through supported system mechanisms, and unregister the application only after verifying it is no longer the default for either scheme. On Windows, if the system picker also changes an independent default, the application SHALL report the mismatch and retain recovery state. Missing or unavailable previous handlers SHALL be reported with a recovery path rather than guessed. On Linux, if a required previous handler cannot be saved and restored, registration SHALL fail before changing defaults. Incomplete changes SHALL retain sufficient state for retry or restoration.

#### Scenario: Undo default-browser setup

- **GIVEN** the previous HTTP/HTTPS handlers were saved and the application is now the default
- **WHEN** the user requests undo after restarting the application
- **THEN** restoration targets the saved handlers separately
- **AND** any required operating-system confirmation is requested
- **AND** the application is unregistered only after verifying that neither scheme still uses it

#### Scenario: The user changed one default independently on Linux

- **GIVEN** a different browser was chosen for HTTPS after registration and HTTP still uses dynamicbrowser
- **WHEN** the user requests undo
- **THEN** only the saved HTTP handler is restored
- **AND** the newer HTTPS choice is preserved

#### Scenario: Only one association changed during registration

- **WHEN** Linux registration changes one scheme but cannot finish or verify the other
- **THEN** the command reports failure and retains the original snapshot
- **AND** the user can retry or restore the scheme that still uses dynamicbrowser

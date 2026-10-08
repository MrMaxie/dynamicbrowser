## ADDED Requirements

### Requirement: Assisted default-browser registration

The application SHALL provide an explicit action to register itself as an available HTTP/HTTPS handler and assist the user in selecting it as the default browser through supported operating-system mechanisms. Startup alone SHALL NOT change the default browser. Windows SHALL be the first supported platform; platform-specific integration SHALL be isolated, and unsupported platforms SHALL report the limitation without claiming success. Protected associations SHALL NOT be overwritten or have their verification bypassed.

#### Scenario: Choose the application as the default browser

- **WHEN** the user requests registration on Windows
- **THEN** the application becomes available as a browser choice
- **AND** the supported default-app settings flow is opened when user confirmation is required
- **AND** success is reported only after the HTTP/HTTPS associations are verified

### Requirement: Restore previous defaults

Before assisting a change to the default browser, the application SHALL retain the previous HTTP and HTTPS handlers separately in persistent state. Repeated registration SHALL NOT replace that snapshot with the application's own associations. An explicit undo action SHALL assist restoration through supported system mechanisms and unregister the application once it is no longer the active handler. Restoration SHALL NOT silently replace a different default chosen by the user since registration. Missing or unavailable previous handlers SHALL be reported with a recovery path rather than guessed.

#### Scenario: Undo default-browser setup

- **GIVEN** the previous HTTP/HTTPS handlers were saved and the application is now the default
- **WHEN** the user requests undo
- **THEN** restoration targets the saved handlers, including after an application restart
- **AND** any required operating-system confirmation is requested
- **AND** the application is unregistered only after it is no longer the default handler

#### Scenario: The user changed the default independently

- **GIVEN** another browser was chosen after the application became the default
- **WHEN** undo is requested
- **THEN** that newer choice is not replaced without explicit confirmation

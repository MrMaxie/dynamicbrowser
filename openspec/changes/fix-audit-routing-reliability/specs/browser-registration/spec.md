## ADDED Requirements

### Requirement: Linux cleanup after manual restoration

Linux `--restore` SHALL clean up its registration without requiring readable recovery data if neither HTTP nor HTTPS still uses DynamicBrowser. If either still uses DynamicBrowser and recovery data is unreadable, it SHALL report the error without replacing the affected handler or removing registration.

#### Scenario: Recovery data is damaged

- **GIVEN** the recovery file is corrupt and the user has already selected another browser for both link types
- **WHEN** the user runs `--restore`
- **THEN** DynamicBrowser removes its own entry and recovery data without replacing either choice

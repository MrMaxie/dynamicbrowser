# Application CLI

## Purpose

Define command-line invocation and version reporting for dynamicbrowser.

## Requirements

### Requirement: Application version

The application SHALL embed its semantic version from the repository's `VERSION` file at build time and print it when invoked with `--version`. Version reporting SHALL exit successfully without starting the tray, changing default-browser associations, or stopping a running instance, including when combined with `--force`.

#### Scenario: Print the version

- **WHEN** the application is invoked with `--version`
- **THEN** standard output is the embedded version followed by a newline
- **AND** the exit status is zero
- **AND** no resident instance is started

#### Scenario: Print the version with force

- **GIVEN** a resident instance is running
- **WHEN** the application is invoked with `--force --version`
- **THEN** the embedded version is printed and the command exits successfully
- **AND** the resident instance remains running

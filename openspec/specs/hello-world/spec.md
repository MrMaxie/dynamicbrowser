# Hello World

## Purpose

Provide the initial Go command-line program for dynamicbrowser.

## Requirements

### Requirement: Greeting

The application SHALL print `Hello, World!` followed by a newline and exit successfully when run without arguments.

#### Scenario: Run the application

- **WHEN** the application is run without arguments
- **THEN** standard output is exactly `Hello, World!` followed by a newline
- **AND** the exit status is zero

### Requirement: Application version

The application SHALL embed its semantic version from the repository's `VERSION` file at build time and print it when invoked with `--version`. The initial version SHALL be `0.1.0`.

#### Scenario: Print the version

- **WHEN** the initial application is invoked with `--version`
- **THEN** standard output is exactly `0.1.0` followed by a newline
- **AND** no greeting is printed
- **AND** the exit status is zero

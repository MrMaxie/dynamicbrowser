## ADDED Requirements

### Requirement: Coexisting desktop targets

The project SHALL support `windows/amd64` and `linux/amd64` as peer desktop targets from one Go module and one `VERSION`. Both SHALL use the same routing semantics, YAML schema, and CLI. Platform-specific configuration locations, desktop integration, and executable paths SHALL be documented without requiring separate configuration formats or source branches. Existing Windows configuration and default-browser setup SHALL remain compatible.

#### Scenario: Configure equivalent routing on both targets

- **GIVEN** equivalent YAML rules and available browser profiles on both targets
- **WHEN** the same HTTP or HTTPS URL and matching inputs are routed
- **THEN** both targets select the same named browser configuration
- **AND** both report the same version when built from the same revision

### Requirement: Native checks and packages

GitHub Actions SHALL check both targets natively on pull requests to the default branch and pushes to it, and support manual runs. Each target SHALL have an independent result for lint, tests, build, and package smoke checks. A successful target SHALL upload its own versioned archive and SHA-256 checksum. The Windows ZIP and artifact name SHALL remain compatible; Linux SHALL produce an executable `tar.gz` archive. Both archives SHALL include the executable, example YAML, YAML schema, README, and license, but no user's configuration. Documentation SHALL cover installation, configuration, setup, restoration, and any desktop prerequisites for each target.

#### Scenario: Build both targets from one revision

- **WHEN** CI builds a revision successfully on both targets
- **THEN** it provides separate Windows and Linux artifacts with the same version
- **AND** each unpacked binary runs `--version` successfully on its native runner
- **AND** a failure of either target cannot be reported as a successful check for that target

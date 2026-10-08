# Approach

- Keep the executable entry point in `cmd/dynamicbrowser` and configuration, routing, tray, autostart, and registration in separate `internal` packages. Windows integration is isolated from the shared routing model.
- Read `config.yaml` beside the executable. Validate known field shapes against the embedded JSON Schema before publishing a typed configuration. Preserve browser declaration order and the last valid snapshot on reload failure; watch the containing directory so atomic replacement is detected.
- Compile ordinary patterns as whole-value globs and `regex:` patterns as Go regular expressions. Evaluate rules in order, AND their specified fields, and OR alternatives within a field. Resolve only available browser executables and invoke them directly with string-array arguments followed by each URL.
- Treat `default` as a fallback, not a prerequisite for matched routing. Never guess a browser from declaration order. Keep a tray override in memory, separate from the configuration.
- Route on demand when no tray owns the singleton. Otherwise hand URLs and source information to the resident. Batch handoff requests within transport limits, retain acknowledged progress, and do not retry completed launches through startup fallback.
- Serialize singleton replacement through a restart gate. Expose shutdown control after tray creation completes, cancel updates before tray removal, and dispatch native menu updates onto their owning Windows thread.
- Register only this application's user-level browser capabilities and activation commands. Pass incoming URLs after `--` so link contents cannot become setup flags. Save separate previous HTTP/HTTPS ProgIDs before publishing registration; do not replace that snapshot on repeated setup.
- Use supported Windows association queries and Default Apps UI, not protected-association writes. Serialize setup operations with a per-user mutex on a COM-initialized, locked thread. Re-query selections after registration and user confirmation before claiming success or removing owned registration.
- Restore only protocols currently owned by dynamicbrowser. If the Settings UI changes an independent protocol as a side effect, report failure and retain recovery state. Remove the saved snapshot last, after owned registration cleanup succeeds.

# Trade-offs

- System selection requires user consent. Windows Settings can change both protocols even from a single link-type picker; verification can detect that side effect but cannot prevent or automatically reverse a protected choice.
- Windows is the supported setup platform. Platform boundaries leave room for later ports without claiming native macOS or Linux verification.
- A repeatable Sandbox harness runs native tests offline. Consent-dependent Settings selection and system-shell activation have a separate manual checklist rather than fragile screen-coordinate automation.

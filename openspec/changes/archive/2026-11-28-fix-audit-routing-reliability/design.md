## Decisions

- Linux uses short, owner-only runtime paths for the sockets and singleton locks. Use the session bus to check tray readiness before starting tray mode, never before opening a URL.
- The Windows routing adapter marks all foreground metadata as a guess, including data handed off by older clients. Configuration, not the sending client, controls whether it is used.
- Apply the identity of the latest valid config independently of its load error. Keep request concurrency bounded and reject excess requests before launching anything.
- Read Linux recovery data only when a current handler still needs restoration. Check both handlers again before removing registration.
- Release validation compares the dates and contents of existing sections as well as their order.

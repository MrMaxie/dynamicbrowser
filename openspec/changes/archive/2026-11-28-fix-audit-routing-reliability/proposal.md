## Why

A newly installed default browser must open links even without a tray. Desktop integration should not switch profiles based on an unreliable source guess or report success or failure for the wrong action.

## What Changes

- Keep Linux link routing independent of tray support and cache paths. Keep sign-in setup warnings separate from link results.
- Make Windows foreground-based source matching opt-in. Keep the last valid configuration active through invalid edits and handle concurrent tray requests without an unbounded queue.
- Clean up Linux registration after a manual restore even if recovery data is corrupt. Preserve published changelog sections in release checks.

## Capabilities

### Modified Capabilities

- `browser-routing`: Clarify source attribution, direct routing and tray handoff.
- `browser-registration`: Allow cleanup when no handler still needs restoration.

## Impact

Windows users relying on source rules must explicitly enable `guess-windows-source` to keep the old heuristic. Destination rules and tray overrides are unchanged.

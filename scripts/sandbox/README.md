# Windows Sandbox checks

Run from the repository on Windows amd64 with Windows Sandbox enabled and the `mise.toml` toolchain installed. Close existing Sandbox windows first; the harness refuses to start alongside another instance.

```powershell
mise exec -- powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Test-WindowsSandbox.ps1
```

The harness copies project source and vendored dependencies to a temporary directory, builds race-enabled test executables, and runs them offline in a fresh Sandbox. It maps that snapshot and the Go SDK read-only; only the result directory is writable. The repository root and installed application configuration are not mapped.

Automated checks cover YAML/schema validation, routing and IPC, singleton replacement, native tray/DPI behavior, live reload, autostart registration, and browser-registration state and cleanup. A final fixture check routes a URL through the built application to a browser probe. These checks do **not** select a default browser in Windows Settings.

The host prints the run directory containing `inputs.json`, `test.wsb`, and `results/`. `inputs.json` records source and binary hashes; `results/result.json` lists completed packages and records `manualChecks: "not run"`. Package logs are UTF-8. A failing or timed-out run raises an error; inspect its logs before retrying. If a timed-out Sandbox remains open, close its window.

## Default Apps and system link activation

Keep the Sandbox open for the consent-dependent checks:

```powershell
mise exec -- powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Test-WindowsSandbox.ps1 -KeepOpen
```

Inside the **Sandbox**, open PowerShell and run:

```powershell
Set-Location C:\Work\bin
.\dynamicbrowser.exe --register
```

1. Select dynamicbrowser for HTTP and HTTPS in Windows Settings, then confirm the application's dialog. Verify both link types on the Settings page.
2. Open these URLs through the system shell:

   ```powershell
   Remove-Item .\browser-launches.jsonl
   Start-Process 'https://example.test/personal/item'
   Start-Process 'https://example.test/work/item'
   Get-Content .\browser-launches.jsonl
   ```

   The probe records `first profile` with the personal URL and `second profile` with the work URL. It records arguments rather than contacting the destination.
3. Start `dynamicbrowser.exe` without arguments. Left-click its tray icon, select `first`, and open the work URL again. The new record must use `first profile`. Choose `Auto` and repeat: it must use `second profile`.
4. Choose `Close tray`, then open the work URL again. On-demand routing must still record `second profile` without reopening the tray.
5. Run `--register` again. With both associations already selected, it must finish without requesting another selection.
6. Run `dynamicbrowser.exe --restore`. Follow the dialog's saved HTTP/HTTPS choices in Windows Settings, then confirm. The application must verify the choices before removing its registration.
7. Repeat setup and cancel its confirmation dialog. The application remains available, but cancellation must not claim successful selection. Choose another browser for both protocols and run `--restore` to clean up.

For mixed-protocol testing, begin with only one protocol assigned to dynamicbrowser and an independent choice for the other. Restoration guidance must mention only the owned protocol. Windows may change both protocols even from a single link-type picker: if the independent choice changes, the application must report a recovery error and retain its registration and saved state, not claim success.

Record manual results separately from `result.json`, including the Windows build and actual protocol choices. Closing the Sandbox discards its registry and application state.

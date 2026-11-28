# Why

dynamicbrowser is distributed and supported on Windows. The repository already has some Linux runtime code, but it does not offer a supported Linux build, installation, or default-browser setup. Add Linux as a second supported desktop target without replacing Windows or creating a separate edition.

# What changes

- Build, check, and distribute `windows/amd64` and `linux/amd64` from the same Go module, default branch, and `VERSION`. Each gets its own native CI job, archive, and checksum.
- Share the routing engine, YAML schema, CLI, and tray behavior. Keep Windows setup and configuration compatible; use per-user XDG integration on Linux.
- On Linux, support explicit HTTP/HTTPS registration and restoration and the same tray features as on Windows on supported desktops, including optional sign-in startup. Routing without a tray remains available on both platforms. Document the verified desktop baseline and the limits of source rules and tray integration.

The Linux target is not supported merely because it compiles: its package, desktop setup, tray, and routing must pass acceptance alongside the existing Windows target.

# Out of scope

Other operating systems or architectures, distro-specific installers, system-wide setup, new routing rule syntax, compositor-specific source capture beyond the Hyprland active-window fallback, and publishing GitHub Releases.

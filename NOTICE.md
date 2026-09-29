HazyVPN is a rebranded, trimmed-down derivative of
[omarchy-vpn](https://github.com/limehawk/omarchy-vpn) by Limehawk, used
under the MIT license (see `LICENSE`).

What carried over: the two-pane Bubble Tea dashboard (Tunnels list +
Details panel), the keybindings (`j/k`, `enter`, `i`, `r`, `x`, `?`, `q`),
the rounded-box rendering, the ANSI-terminal color theme, and the
security-conscious root helper design (a single allow-listed helper binary
invoked via a narrow sudoers rule, name/path validation, hook-command
refusal on import).

What changed for HazyVPN: renamed throughout (binary, helper path, log/state
paths, sudoers file); NetBird, Cloudflare WARP, Waybar/Quickshell desktop-bar
integration, and the Arch/AUR-specific packaging were removed to keep this a
plain, distro-agnostic WireGuard tunnel manager; the sudoers rule targets the
installing user directly instead of the `wheel` group, so it works the same
on Debian/Ubuntu-family systems (which use `sudo`, not `wheel`) as it does on
Arch; install/uninstall scripts and a Makefile replace the PKGBUILD.

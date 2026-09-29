# HazyVPN

A fast, keyboard-driven TUI for importing and managing your WireGuard
tunnels — connect, disconnect, import, rename, and delete `.conf` files
without leaving the terminal.

```
 HazyVPN                                                                dev
 Connected  homelab

 ╭─ Tunnels ──────────────────────────────────╮  ╭─ Details ─────────────────────────────────────╮
 │                                            │  │                                                │
 │   homelab                                  │  │   Status       Connected                      │
 │   WireGuard · Connected                    │  │   Address      10.8.0.2/24                    │
 │                                            │  │   Endpoint     vpn.example.net:51820           │
 │   office                                   │  │   Download     12.40 MiB received              │
 │   WireGuard · Disconnected                 │  │   Upload       3.10 MiB sent                   │
 │                                            │  │   Handshake    1 minute ago                    │
 │   travel                                   │  │   Path         /etc/wireguard/homelab.conf     │
 │   WireGuard · Disconnected                 │  │                                                │
 │                                            │  │                                                │
 ╰────────────────────────────────────────────╯  ╰────────────────────────────────────────────────╯

 enter toggle • i import • ? help • q quit
```

Run `hazyvpn --demo` to see this exact screen with fake sample tunnels —
no root, no `wireguard-tools`, no real config required.

## Keybindings

| Key       | Action                    |
|-----------|----------------------------|
| `j` / `↓` | move down                 |
| `k` / `↑` | move up                   |
| `enter`   | connect / disconnect      |
| `i`       | import a `.conf`/`.wg` file |
| `r`       | rename the selected config |
| `x`       | delete the selected config  |
| `?`       | help                      |
| `q`       | quit                      |
| `ctrl+c`  | force quit                |

## Install

Requires Linux with `wireguard-tools`. The installer will fetch it (and a
bootstrap Go toolchain, if needed) via `apt`, `pacman`, `dnf`, or `zypper`.

```sh
git clone <this-repo-url> hazyvpn && cd hazyvpn
./install.sh
```

or, with `make`:

```sh
make install
```

This builds the binary, installs it to `/usr/local/bin/hazyvpn`, installs a
root helper to `/usr/local/lib/hazyvpn/helper`, and adds a narrow sudoers
rule so **your user only** can run that one helper without a password
prompt — nothing else gains passwordless root.

To uninstall (your `/etc/wireguard` configs are never touched):

```sh
./uninstall.sh   # or: make uninstall
```

### Deploying to another machine

Copy this directory to the target machine (`git clone` from your own
remote, `scp -r`, `rsync`, or a tarball) and run `./install.sh` there. The
script is distro-agnostic — it works the same on Debian/Ubuntu, Arch, and
Fedora/openSUSE family systems.

## How it stays safe

WireGuard tunnels live in `/etc/wireguard` and require root to bring up or
down. Rather than giving the TUI (or a whole sudo group) blanket root
access, HazyVPN talks to a single, narrow helper script:

- The sudoers rule allows exactly one command, for exactly the user who ran
  `install.sh` — `NOPASSWD: /usr/local/lib/hazyvpn/helper`.
- The helper takes **config names, never paths**, validates them against a
  strict character set, and builds every `/etc/wireguard/...` path itself.
- Private and preshared keys are stripped before a config is ever shown to
  the TUI.
- Imports are refused if the config contains `PreUp`/`PostUp`/`PreDown`/
  `PostDown` hooks, since those run as root on every connect — the TUI will
  instead prompt for your password once and install it with plain `sudo`.
- Every connect/disconnect/import/rename/delete is logged to the system
  journal (`journalctl -t hazyvpn`) and to `~/.local/state/hazyvpn/hazyvpn.log`.

## Development

```sh
make build   # go build
make test    # go test ./...
make vet     # go vet ./...
make demo    # build and run with fake sample tunnels
```

Built with [Bubble Tea](https://charm.land) (v2), Bubbles, and Lip Gloss.

## Credits

HazyVPN is a rebranded, distro-agnostic derivative of
[omarchy-vpn](https://github.com/limehawk/omarchy-vpn) by Limehawk (MIT
license). See `NOTICE.md` for exactly what carried over and what changed.

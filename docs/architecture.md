# TV Remote Session: architecture

Remote setup and control of customers' Android TV devices (NVIDIA Shield, Google TV, Fire TV)
from anywhere, driven by a CLI that a person or an AI agent can use.

## Pieces

```
Customer TV                                   Operator machine (NAS / PC)
+----------------------------------+          +-------------------------------+
| TV app (Kotlin)                  |          | tvctl (Go, single binary)     |
|  - embedded Tailscale (tsnet)    | tailnet  |  - device registry            |
|  - forwards tailnet:5555 to the  |<-------->|  - adb wrapper, fan-out       |
|    TV's own adbd                 |          |  - UI tree for agents         |
|  - setup guide, consent, Stop    |          |  - scripts (YAML)             |
|  - boot + update restart         |          |  - scrcpy for live view       |
+----------------------------------+          +-------------------------------+
```

- **Transport:** Tailscale. No relay server of our own; Tailscale's DERP relays cover routers
  that block direct connections. The TV app embeds Tailscale so the customer never sees it.
- **Control:** plain ADB over the tailnet. Everything (keys, text, installs, settings,
  screenshots, UI tree, reboot) is an ADB command, so there is no custom protocol to maintain.
- **Live view:** scrcpy over the same ADB connection.

## Customer onboarding (one time)

1. Install Downloader from the store, type our short code, install the TV app.
2. Open the app: it joins the tailnet and shows "Connected".
3. Unlock Developer Options and turn on ADB debugging (the app shows each click; the
   operator can watch through a screen-share prompt).
4. Tap "Always allow" on the first ADB connection.

Steps 1 and 3 are required by Amazon and Google and cannot be automated.

## Consent rules (non-negotiable)

- The app has a **Stop remote help** button that disconnects and blocks reconnects until the
  customer turns it back on.
- Optional per-session **Allow?** prompt (default on, "always allow" opt-in).
- The app re-enables ADB after a reboot or update only if the customer has not turned it off
  themselves. A customer's off switch is never overridden.
- Uninstalling the app offers to turn ADB off.
- tailnet ACLs: customer TVs cannot reach each other or the operator's LAN; only tagged
  operator machines can reach TV port 5555.

## tvctl design

- Targets: a device name, `--group <g>`, or `--all`. Fan-out runs devices in parallel with a
  concurrency limit; each device reports its own result.
- Output: human text by default, `--json` for agents (one JSON document per command).
- `ui` returns a compact list of on-screen elements (text, id, bounds, focused, clickable),
  which is what an agent reads to decide the next key press or click.
- Registry: `~/.config/tvctl/devices.json` (names, addresses, groups). No secrets in it; the
  ADB key stays in `~/.android/adbkey` and is never printed.

## Build order

1. tvctl backend against real devices on the LAN (no TV app needed yet).
2. TV app: tsnet join + adb forward + boot/update restart.
3. TV app onboarding UI (sample/mockup first, approved before building).
4. Optional operator web dashboard served by tvctl.

# TV Remote Session: architecture

Remote setup and control of Android TV devices (NVIDIA Shield, Google TV, Fire TV)
from anywhere, driven by a CLI that a person or an AI agent can use. No server to
run or pay for.

## Pieces

```
Customer home                                  Helper's machine
+----------------------------------+           +-------------------------------+
| tvlink (TV app, or phone app     |  direct   | tvctl (Go, single binary)     |
|  for Fire TV)                    |  WebRTC   |  - tvctl link: tunnel to TV   |
|  - answers helpers with the code |<--------->|  - device registry            |
|  - forwards each stream to the   |           |  - adb wrapper, fan-out       |
|    TV's adb (port 5555)          |           |  - UI tree for agents         |
|  - setup guide, consent, Stop    |           |  - scripts (YAML)             |
+----------------------------------+           +-------------------------------+
              \                                   /
               \---- public Nostr relays --------/
                  (pass two sealed setup notes)
```

- **Control:** plain ADB. Keys, text, installs, settings, screenshots, UI tree and
  reboot are all ADB commands, so there is no custom protocol to maintain.
- **Transport:** a direct WebRTC connection between the two ends, set up through
  free public services only (see below). `tvctl link` serves the TV's adb on a
  loopback port, and every other tvctl command uses it like a LAN TV.
- **Live view:** scrcpy over the same ADB connection.

## How the two ends find each other

1. The TV side creates a random 128-bit pairing code and stores it (mode 0600).
   The code never leaves the device except when the customer shares it.
2. Both ends derive from the code a topic (hash) and an AES-GCM key (HKDF).
3. The helper posts an offer note, sealed with that key, to five public Nostr
   relays under the topic, using a throwaway Nostr key. The TV watches the topic,
   opens the note, and posts a sealed answer. Notes older than 2 minutes are
   ignored, and notes sealed with another code fail to open.
4. Each note carries the sender's network addresses, found with free public STUN
   servers (Google, Cloudflare). WebRTC then connects the two ends directly,
   encrypted with DTLS whose fingerprints came inside the sealed notes.
5. Each adb TCP connection becomes its own data channel, with backpressure so
   large installs do not fill memory.

The relays only ever see opaque sealed text. Some strict networks (symmetric NAT
on both ends) cannot connect directly; those need a TURN relay. Planned: the
helper's own TURN (Cloudflare's free tier, or a self-hosted coturn) as an option.

## Connection modes

| Mode | When | How |
|---|---|---|
| Direct link (default) | Anywhere | `tvctl link --code-file <file>`, then `tvctl add <name> 127.0.0.1:15555` |
| Same network | Helper on the TV's LAN | `tvctl add <name> <tv-ip>:5555` |
| Own VPN | Helper and TV on the helper's own Tailscale/WireGuard network | `tvctl add <name> <vpn-ip>:5555` |

Each helper uses their own code, relays, and (if any) VPN or TURN. Nothing is
shared with the project author.

## Where the bridge runs

| Device | Bridge | Why |
|---|---|---|
| NVIDIA Shield, Google TV | On the TV | adb accepts on-device connections to 127.0.0.1:5555 (checked on Shield, Android 11; full link test pending) |
| Fire TV | On a phone or other device on the same Wi-Fi | Fire OS refuses on-device adb (see [fire-tv.md](fire-tv.md)) |

## Customer onboarding (one time)

1. Turn on ADB debugging (Developer options; on Fire TV press the device name in
   About 7 times first).
2. Install the app (TV, or phone for Fire TV) and share the pairing code.
3. Tick "Always allow" and press OK on the first ADB connection.

The ADB switch and the first approval are required by Google and Amazon and
cannot be automated.

## Consent rules (non-negotiable)

- The app has a **Stop remote help** button that disconnects and blocks reconnects
  until the customer turns it back on.
- A per-session **Allow?** prompt (default on, "always allow" opt-in).
- A customer's ADB off switch is never overridden.
- A new pairing code can be created at any time, which locks out every helper
  holding the old one.

## tvctl design

- Targets: a device name, `--group <g>`, or `--all`. Fan-out runs devices in
  parallel with a concurrency limit; each device reports its own result.
- Output: human text by default, `--json` for agents.
- `ui` returns a compact list of on-screen elements (text, id, bounds, focused,
  clickable), which is what an agent reads to decide the next key press or click.
- Registry: `~/.config/tvctl/devices.json` (names, addresses, groups). No secrets
  in it; the ADB key stays in `~/.android/adbkey` and is never printed.

## Status

- Done: tvctl against LAN TVs; direct link tested from a remote server to a Fire
  TV Stick through a LAN bridge (2026-10-18).
- Next: TV/phone app (Kotlin) with consent screens, short codes, optional TURN.
